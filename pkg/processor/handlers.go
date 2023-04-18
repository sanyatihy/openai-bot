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
	case strings.EqualFold(text, "/clear"):
		return p.handleClearCommand(ctx, message)
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
			"/clear - Clear conversation context\n" +
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
		Text: fmt.Sprintf("I send your messages to OpenAI GPT.\n"+
			"Current GPT model: %s", response.ID),
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func (p *processor) handleClearCommand(ctx context.Context, message telegram.Message) error {
	err := p.db.ClearChatContext(ctx, message.Chat.ID)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error clearing chat %d context in db", message.Chat.ID), zap.Error(err))
		return err
	}

	_, err = p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: message.Chat.ID,
		Text:   "Conversation context successfully cleared.",
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

	existingContext, err := p.db.GetChatContext(ctx, message.Chat.ID)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error retrieving chat %d context from db", message.Chat.ID), zap.Error(err))
		return err
	}

	var messages []openai.Message
	if existingContext != nil {
		messages = existingContext
	}
	messages = append(messages, openai.Message{
		Role:    "user",
		Content: text,
	})

	response, err := p.openAIClient.ChatCompletion(ctx, &openai.ChatCompletionRequest{
		Model:     openAIModelID,
		Messages:  messages,
		N:         1,
		Stream:    false,
		MaxTokens: 2048,
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

	messages = append(messages, openai.Message{
		Role:    "assistant",
		Content: response.Choices[0].Message.Content,
	})

	err = p.db.UpdateChatContext(ctx, message.Chat.ID, messages)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error updating chat %d context in db", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}
