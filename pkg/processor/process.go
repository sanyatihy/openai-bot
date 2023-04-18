package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"go.uber.org/zap"
)

const (
	openAIModelID = "gpt-3.5-turbo"
)

var (
	pricingPerOneK = map[string]float64{
		"gpt-3.5-turbo": 0.0002,
	}
)

func (p *processor) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := p.db.RunInitialMigrations(ctx)
	if err != nil {
		p.logger.Error("Failed to run initial migrations", zap.Error(err))
		return err
	}

	go p.GetUpdates()
	go p.ProcessUpdates(p.processUpdate)

	return nil
}

func (p *processor) GetUpdates() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		var lastUpdateID int
		err := p.RetryWithBackoff(3, func() error {
			var err error
			lastUpdateID, err = p.queue.GetLastChatUpdate(ctx)
			if err != nil {
				p.logger.Error("Error", zap.Error(err))
			}
			return err
		})
		if err != nil {
			p.logger.Error("Failed to load last update ID", zap.Error(err))
		}

		getUpdatesRequest := &telegram.GetUpdatesRequest{
			Offset:  lastUpdateID + 1,
			Timeout: 30,
		}

		p.logger.Info("Getting updates...")

		var updates []telegram.Update
		err = p.RetryWithBackoff(5, func() error {
			updates, err = p.tgBotClient.GetUpdates(ctx, getUpdatesRequest)
			if err != nil {
				p.logger.Error("Error", zap.Error(err))
			}
			return err
		})
		if err != nil {
			p.logger.Error("Error getting updates", zap.Error(err))
		}

		for _, update := range updates {
			err := p.RetryWithBackoff(3, func() error {
				var err error
				err = p.queue.InsertChatUpdate(ctx, update)
				if err != nil {
					p.logger.Error("Error", zap.Error(err))
				}
				return err
			})
			if err != nil {
				p.logger.Error("Failed to insert chat updates", zap.Error(err))
			}
		}

		time.Sleep(3 * time.Second)
		cancel()
	}
}

func (p *processor) ProcessUpdates(processUpdate func(ctx context.Context, update telegram.Update) error) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		updateID, update, err := p.queue.GetNextChatUpdate(ctx, "processing")
		if err != nil {
			p.RetryWithBackoff(2, func() error {
				updateID, update, err = p.queue.GetNextChatUpdate(ctx, "processing")
				if err != nil {
					p.logger.Error("Error", zap.Error(err))
				}
				return err
			})
			if err != nil {
				p.logger.Error("Error", zap.Error(err))
				continue
			}
		}

		if updateID == 0 {
			time.Sleep(3 * time.Second)
			continue
		}

		p.queueSemaphore <- struct{}{}

		go func(updateID int, update telegram.Update) {
			defer func() { <-p.queueSemaphore }()
			ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)

			err := processUpdate(ctx, update)
			if err != nil {
				_ = p.queue.SetChatUpdateStatus(ctx, updateID, "error")
			} else {
				_ = p.queue.SetChatUpdateStatus(ctx, updateID, "processed")
			}

			cancel()
		}(updateID, update)

		cancel()
	}
}

func (p *processor) processUpdate(ctx context.Context, update telegram.Update) error {
	if update.Message.Text == nil {
		p.logger.Error("Error, got empty message text")

		_, err := p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
			ChatID: update.Message.Chat.ID,
			Text:   "That doesn't look like a valid message to me, try again",
		})
		if err != nil {
			p.logger.Error(fmt.Sprintf("Error sending message to chat %d", update.Message.Chat.ID), zap.Error(err))
			return err
		}
	}

	if strings.HasPrefix(*update.Message.Text, "/") {
		if err := p.handleCommand(ctx, update.Message); err != nil {
			p.logger.Error(fmt.Sprintf("Error handling command in chat %d", update.Message.Chat.ID), zap.Error(err))
			return err
		}
	} else {
		if err := p.handleMessage(ctx, update.Message); err != nil {
			p.logger.Error(fmt.Sprintf("Error handling message in chat %d", update.Message.Chat.ID), zap.Error(err))
			return err
		}
	}

	return nil
}
