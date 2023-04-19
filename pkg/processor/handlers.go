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

func (p *processor) sendMessage(ctx context.Context, chatID int, text string) error {
	_, err := p.tgBotClient.SendMessage(ctx, &telegram.SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	})
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", chatID), zap.Error(err))
	}
	return err
}

func (p *processor) handleStartCommand(ctx context.Context, message telegram.Message) error {
	text := "Welcome to the bot!"
	return p.sendMessage(ctx, message.Chat.ID, text)
}

func (p *processor) handleHelpCommand(ctx context.Context, message telegram.Message) error {
	text := `Here is a list of available commands:

/start - Start the bot
/clear - Clear conversation context
/help - Show help message
/about - About the bot
`
	return p.sendMessage(ctx, message.Chat.ID, text)
}

func (p *processor) handleAboutCommand(ctx context.Context, message telegram.Message) error {
	response, err := p.openAIClient.GetModel(ctx, openAIModelID)
	if err != nil {
		return err
	}

	text := fmt.Sprintf("I send your messages to OpenAI GPT.\nCurrent GPT model: %s", response.ID)
	return p.sendMessage(ctx, message.Chat.ID, text)
}

func (p *processor) handleClearCommand(ctx context.Context, message telegram.Message) error {
	err := p.db.ClearChatContext(ctx, message.Chat.ID)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error clearing chat %d context in db", message.Chat.ID), zap.Error(err))
		return err
	}

	text := "Conversation context successfully cleared."
	return p.sendMessage(ctx, message.Chat.ID, text)
}

func (p *processor) handleUnknownCommand(ctx context.Context, message telegram.Message) error {
	text := "Sorry, I didn't understand that command. Type /help for a list of available commands."
	return p.sendMessage(ctx, message.Chat.ID, text)
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

	err = p.sendMessage(ctx, message.Chat.ID, response.Choices[0].Message.Content)
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
