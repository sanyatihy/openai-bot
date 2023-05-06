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

type UserSettings struct {
	UserID   int
	GPTModel string
}

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
	case strings.EqualFold(text, "/settings"):
		return p.handleSettingsCommand(ctx, message)
	default:
		return p.handleUnknownCommand(ctx, message)
	}
}

func (p *processor) sendMessage(ctx context.Context, chatID int, text string, replyMarkup *telegram.InlineKeyboardMarkup) error {
	req := &telegram.SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}

	if replyMarkup != nil {
		req.ReplyMarkup = replyMarkup
	}

	_, err := p.tgBotClient.SendMessage(ctx, req)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", chatID), zap.Error(err))
	}
	return err
}

func (p *processor) handleStartCommand(ctx context.Context, message telegram.Message) error {
	text := "Welcome to the bot!"
	return p.sendMessage(ctx, message.Chat.ID, text, nil)
}

func (p *processor) handleHelpCommand(ctx context.Context, message telegram.Message) error {
	text := `Here is a list of available commands:

/start - Start the bot
/clear - Clear conversation context
/help - Show help message
/about - About the bot
`
	return p.sendMessage(ctx, message.Chat.ID, text, nil)
}

func (p *processor) handleAboutCommand(ctx context.Context, message telegram.Message) error {
	text := fmt.Sprintf("I send your messages to OpenAI API")
	return p.sendMessage(ctx, message.Chat.ID, text, nil)
}

func (p *processor) handleClearCommand(ctx context.Context, message telegram.Message) error {
	err := p.db.ClearChatContext(ctx, message.Chat.ID)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error clearing chat %d context in db", message.Chat.ID), zap.Error(err))
		return err
	}

	text := "Conversation context successfully cleared."
	return p.sendMessage(ctx, message.Chat.ID, text, nil)
}

func (p *processor) handleSettingsCommand(ctx context.Context, message telegram.Message) error {
	settingsMenu := p.generateSettingsMenu()
	text := "Update settings:"
	return p.sendMessage(ctx, message.Chat.ID, text, settingsMenu)
}

func (p *processor) handleUnknownCommand(ctx context.Context, message telegram.Message) error {
	text := "Sorry, I didn't understand that command. Type /help for a list of available commands."
	return p.sendMessage(ctx, message.Chat.ID, text, nil)
}

func (p *processor) handleMessage(ctx context.Context, message telegram.Message) error {
	if message.Text == nil {
		p.logger.Error("Error, got empty message text")
		return &InternalError{
			Message: "error, got empty message text",
		}
	}
	text := *message.Text

	modelID, existingContext, err := p.db.GetChatContext(ctx, message.Chat.ID)
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

	var model string
	if modelID != "" {
		model = modelID
	} else {
		model = openAIModelID["gpt-4"]
	}

	response, err := p.openAIClient.ChatCompletion(ctx, &openai.ChatCompletionRequest{
		Model:     model,
		Messages:  messages,
		N:         1,
		Stream:    false,
		MaxTokens: 2048,
	})
	if err != nil {
		return err
	}

	cost := float64(response.Usage.PromptTokens)*pricingPerOneK[model]["prompt"]/1024 + float64(response.Usage.CompletionTokens)*pricingPerOneK[model]["completion"]/1024
	p.logger.Info(fmt.Sprintf("Got chat completion response, tokens used: %d, cost: %.5f$", response.Usage.TotalTokens, cost))

	messageText := fmt.Sprintf("%s\n\nModel: %s, Tokens used: %d, Cost: %.5f$", response.Choices[0].Message.Content, model, response.Usage.TotalTokens, cost)
	err = p.sendMessage(ctx, message.Chat.ID, messageText, nil)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error sending message to chat %d", message.Chat.ID), zap.Error(err))
		return err
	}

	messages = append(messages, openai.Message{
		Role:    "assistant",
		Content: response.Choices[0].Message.Content,
	})

	err = p.db.UpdateChatContext(ctx, message.Chat.ID, messages, model)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Error updating chat %d context in db", message.Chat.ID), zap.Error(err))
		return err
	}

	return nil
}

func (p *processor) handleCallbackQuery(ctx context.Context, callbackQuery *telegram.CallbackQuery) error {
	var modelID string

	switch callbackQuery.Data {
	case "gpt_model":
		gptModelMenu := p.generateGPTModelMenu()
		text := "Set GPT model:"
		return p.sendMessage(ctx, callbackQuery.Message.Chat.ID, text, gptModelMenu)
	case "gpt_3_5":
		modelID = openAIModelID["gpt-3.5"]
	case "gpt_4":
		modelID = openAIModelID["gpt-4"]
	default:
		return p.handleUnknownCommand(ctx, *callbackQuery.Message)
	}

	err := p.db.UpdateChatModel(ctx, callbackQuery.Message.Chat.ID, modelID)
	if err != nil {
		p.logger.Error(fmt.Sprintf("Failed to update chat %d gpt model in db", callbackQuery.Message.Chat.ID), zap.Error(err))
		return err
	}

	text := fmt.Sprintf("GPT model set to %s", modelID)
	return p.sendMessage(ctx, callbackQuery.Message.Chat.ID, text, nil)
}

func (p *processor) generateSettingsMenu() *telegram.InlineKeyboardMarkup {
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "GPT model", CallbackData: "gpt_model"},
			},
		},
	}
}

func (p *processor) generateGPTModelMenu() *telegram.InlineKeyboardMarkup {
	return &telegram.InlineKeyboardMarkup{
		InlineKeyboard: [][]telegram.InlineKeyboardButton{
			{
				{Text: "gpt-3.5", CallbackData: "gpt_3_5"},
				{Text: "gpt-4", CallbackData: "gpt_4"},
			},
		},
	}
}
