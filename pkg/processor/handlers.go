package processor

import (
	"context"
	"fmt"
	"strings"

	"github.com/sanyatihy/openai-bot/pkg/telegram"
	"github.com/sanyatihy/openai-go/pkg/openai"
	"go.uber.org/zap"
)

type Handler func(message telegram.Message) error

func (p *processor) handleCommand(ctx context.Context, message telegram.Message) error {
	if message.Text == nil {
		p.logger.Error("Error, got empty message text")
		return &InternalError{
			Message: "error, got empty message text",
		}
	}
	text := *message.Text

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
		Text: "Here is a list of available commands:\n\n" +
			"/start - Start the bot\n" +
			"/help - Show help message\n" +
			"/about - About the bot",
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func (p *processor) handleAboutCommand(ctx context.Context, message telegram.Message) error {
	response, err := p.openAIClient.GetModel(ctx, openAIModelID)
	if err != nil {
		return err
	}

	_, err = p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text: fmt.Sprintf("I transfer your messages to OpenAI GPT.\n"+
			"I don't support coversation context yet.\n"+
			"One question - one answer.\n"+
			"Current GPT model: %s", response.ID),
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
	if message.Text == nil {
		p.logger.Error("Error, got empty message text")
		return &InternalError{
			Message: "error, got empty message text",
		}
	}
	text := *message.Text

	response, err := p.openAIClient.ChatCompletion(ctx, &openai.ChatCompletionRequest{
		Model: openAIModelID,
		Messages: []openai.Message{
			{
				Role:    "user",
				Content: text,
			},
		},
		N:         1,
		Stream:    false,
		MaxTokens: 1024,
	})
	if err != nil {
		return err
	}

	price := float64(response.Usage.TotalTokens) * pricingPerOneK[openAIModelID] / 1024
	p.logger.Info(fmt.Sprintf("Got chat completion response, tokens used: %d, price: %f", response.Usage.TotalTokens, price))

	_, err = p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   response.Choices[0].Message.Content,
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}
