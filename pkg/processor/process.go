package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sanyatihy/openai-bot/pkg/storage"
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

type updateWithID struct {
	updateID int
	update   telegram.Update
}

func (p *processor) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	p.initWorkers()

	err := p.db.RunInitialMigrations(ctx)
	if err != nil {
		p.logger.Error("Failed to run initial migrations", zap.Error(err))
		return err
	}

	go p.getUpdates()
	go p.processUpdates()
	go p.cleanupProcessingUpdates()

	return nil
}

func (p *processor) initWorkers() {
	for i := 0; i < p.concurrentWorkers; i++ {
		go p.worker(i)
	}
}

func (p *processor) getUpdates() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		var lastUpdateID int
		err := p.RetryWithBackoff(3, func() error {
			var err error
			lastUpdateID, err = p.queue.GetLastChatUpdateID(ctx)
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

		p.insertUpdates(ctx, updates)

		time.Sleep(3 * time.Second)
		cancel()
	}
}

func (p *processor) insertUpdates(ctx context.Context, updates []telegram.Update) {
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
}

func (p *processor) worker(id int) {
	for updateWithID := range p.queueUpdates {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		err := p.processUpdate(ctx, updateWithID.update)
		status := storage.UpdateStatusProcessed
		if err != nil {
			status = storage.UpdateStatusError
		}
		err = p.RetryWithBackoff(3, func() error {
			err = p.queue.SetChatUpdateStatus(ctx, updateWithID.updateID, status)
			if err != nil {
				p.logger.Error("Error", zap.Error(err))
			}
			return err
		})
		if err != nil {
			p.logger.Error("Error", zap.Error(err))
		}
		cancel()
	}
}

func (p *processor) processUpdates() {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		var updateID int
		var update telegram.Update
		err := p.RetryWithBackoff(3, func() error {
			var err error
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

		if updateID == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		p.queueUpdates <- updateWithID{updateID: updateID, update: update}

		cancel()
	}
}

func (p *processor) processUpdate(ctx context.Context, update telegram.Update) error {
	if update.Message.Text == nil {
		p.logger.Error("Got empty message text")

		_, err := p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
			ChatID: update.Message.Chat.ID,
			Text:   "That doesn't look like a valid message to me, try again",
		})
		if err != nil {
			p.logger.Error(fmt.Sprintf("Error sending message to chat %d", update.Message.Chat.ID), zap.Error(err))
			return err
		}

		return &InternalError{
			Message: "got empty message text",
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

func (p *processor) cleanupProcessingUpdates() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			p.logger.Info("Cleaning up stuck updates...")
			err := p.queue.ResetChatUpdatesStatus(ctx)
			if err != nil {
				p.logger.Error("Error", zap.Error(err))
			}
			cancel()
		}
	}
}
