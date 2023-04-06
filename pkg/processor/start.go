package processor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"go.uber.org/zap"
	"os"
	"time"
)

const lastUpdateIDFile = "last_update_id.json"

func (p *processor) Start(ctx context.Context) error {
	response, err := p.openAIClient.GetModel(ctx, "gpt-3.5-turbo")
	if err != nil {
		p.logger.Error("Error:", zap.Error(err))
	}
	p.logger.Info(fmt.Sprintf("Got response: %s", response.ID))

	lastUpdateID, err := p.loadLastUpdateIDFromFile(lastUpdateIDFile)
	if err != nil {
		p.logger.Error("Error loading last update ID: ", zap.Error(err))
	}

	for {
		requestOptions := &telegram.GetUpdatesRequest{
			Offset: lastUpdateID + 1,
		}
		updates, err := p.tgBotClient.GetUpdates(ctx, requestOptions)
		if err != nil {
			p.logger.Error("Error getting updates:", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}

		for _, update := range updates {
			if update.UpdateID > lastUpdateID {
				lastUpdateID = update.UpdateID
			}

			p.logger.Info(fmt.Sprintf("Got text: %s", update.Message.Text))

			requestOptions := &telegram.SendMessageRequest{
				ChatID: update.Message.Chat.ID,
				Text:   update.Message.Text,
			}
			_, err = p.tgBotClient.SendMessage(ctx, requestOptions)
			if err != nil {
				p.logger.Error(fmt.Sprintf("Error sending message to chat %d:", update.Message.Chat.ID), zap.Error(err))
			}
		}

		err = p.saveLastUpdateIDToFile(lastUpdateIDFile, lastUpdateID)
		if err != nil {
			p.logger.Error("Error saving last update ID: ", zap.Error(err))
		}

		time.Sleep(5 * time.Second)
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
