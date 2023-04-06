package processor

import (
	"context"
	"fmt"
	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"go.uber.org/zap"
	"strings"
)

type Handler func(message telegram.Message) error

func (p *processor) handleCommand(ctx context.Context, message telegram.Message) error {
	text := message.Text

	switch {
	case strings.EqualFold(text, "/start"):
		return p.handleStartCommand(ctx, message)
	case strings.EqualFold(text, "/help"):
		return p.handleHelpCommand(ctx, message)
	case strings.EqualFold(text, "/about"):
		return p.handleAboutCommand(ctx, message)
	default:
		return p.handleUnknownCommand(ctx, message)
	}
}

func (p *processor) handleStartCommand(ctx context.Context, message telegram.Message) error {
	_, err := p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   "Welcome to the bot!",
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func (p *processor) handleHelpCommand(ctx context.Context, message telegram.Message) error {
	_, err := p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   "Here is a list of available commands:\n/start - Start the bot\n/help - Show help message",
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func (p *processor) handleAboutCommand(ctx context.Context, message telegram.Message) error {
	response, err := p.openAIClient.GetModel(ctx, "gpt-3.5-turbo")
	if err != nil {
		return err
	}

	_, err = p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   fmt.Sprintf("Current GPT model ID: %s", response.ID),
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func (p *processor) handleUnknownCommand(ctx context.Context, message telegram.Message) error {
	_, err := p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   "Sorry, I didn't understand that command. Type /help for a list of available commands.",
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func (p *processor) handleMessage(ctx context.Context, message telegram.Message) error {
	_, err := p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   message.Text,
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}
