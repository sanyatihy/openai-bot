package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

const lastUpdateIDFile = "last_update_id.json"

func (p *processor) Start() error {
	lastUpdateID, err := p.loadLastUpdateIDFromFile(lastUpdateIDFile)
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

			if strings.HasPrefix(update.Message.Text, "/") {
				if err = p.handleCommand(ctx, update.Message); err != nil {
					p.logger.Error(fmt.Sprintf("Error handling command in chat %d", update.Message.Chat.ID), zap.Error(err))
				}
			} else {
				if err = p.handleMessage(ctx, update.Message); err != nil {
					p.logger.Error(fmt.Sprintf("Error handling message in chat %d", update.Message.Chat.ID), zap.Error(err))
				}
			}
		}

		err = p.saveLastUpdateIDToFile(lastUpdateIDFile, lastUpdateID)
		if err != nil {
			p.logger.Error("Error saving last update ID", zap.Error(err))
		}

		time.Sleep(3 * time.Second)
		cancel()
	}
}

func (p *processor) loadLastUpdateIDFromFile(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	defer file.Close()

	data := make(map[string]int)
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		return 0, err
	}

	return data["lastUpdateID"], nil
}

func (p *processor) saveLastUpdateIDToFile(filename string, lastUpdateID int) error {
	data := map[string]int{"lastUpdateID": lastUpdateID}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(data)
}
