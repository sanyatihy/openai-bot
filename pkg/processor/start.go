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
	openAIModelID    = "gpt-3.5-turbo"
	lastUpdateIDFile = "last_update_id.json"
)

var (
	pricingPerOneK = map[string]float64{
		"gpt-3.5-turbo": 0.0002,
	}
)

func (p *processor) Start() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := p.runInitialMigrations(ctx)
	if err != nil {
		p.logger.Error("Failed to run initial migrations", zap.Error(err))
		return err
	}

	lastUpdateID, err := p.loadLastUpdateIDFromDB(ctx)
	if err != nil {
		p.logger.Error("Error loading last update ID", zap.Error(err))
	}

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)

		getUpdatesRequest := &telegram.GetUpdatesRequest{
			Offset:  lastUpdateID + 1,
			Timeout: 30,
		}

		p.logger.Info("Getting updates...")

		var updates []telegram.Update
		err := p.RetryWithBackoff(5, func() error {
			var err error
			updates, err = p.tgBotClient.GetUpdates(ctx, getUpdatesRequest)
			return err
		})
		if err != nil {
			p.logger.Error("Error getting updates", zap.Error(err))
		}

		for _, update := range updates {
			if update.UpdateID > lastUpdateID {
				lastUpdateID = update.UpdateID
			}

			if update.Message.Text == nil {
				p.logger.Error("Error, got empty message text")

				_, err = p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
					ChatID: update.Message.Chat.ID,
					Text:   "That doesn't look like a valid message to me, try again",
				})
				if err != nil {
					p.logger.Error(fmt.Sprintf("Error sending message to chat %d", update.Message.Chat.ID), zap.Error(err))
				}

				continue
			}

			if strings.HasPrefix(*update.Message.Text, "/") {
				if err = p.handleCommand(ctx, update.Message); err != nil {
					p.logger.Error(fmt.Sprintf("Error handling command in chat %d", update.Message.Chat.ID), zap.Error(err))
				}
			} else {
				if err = p.handleMessage(ctx, update.Message); err != nil {
					p.logger.Error(fmt.Sprintf("Error handling message in chat %d", update.Message.Chat.ID), zap.Error(err))
				}
			}
		}

		err = p.saveLastUpdateIDToDB(ctx, lastUpdateID)
		if err != nil {
			p.logger.Error("Error saving last update ID", zap.Error(err))
		}

		time.Sleep(3 * time.Second)
		cancel()
	}
}
