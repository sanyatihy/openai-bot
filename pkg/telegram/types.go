package telegram

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       Message        `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query,omitempty"`
}

type Message struct {
	MessageID   int                   `json:"message_id"`
	Text        *string               `json:"text,omitempty"`
	Chat        Chat                  `json:"chat"`
	From        *User                 `json:"user,omitempty"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type Chat struct {
	ID int `json:"id"`
}

type SendMessageRequest struct {
	ChatID      int                   `json:"chat_id"`
	Text        string                `json:"text"`
	ReplyMarkup *InlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type GetUpdatesRequest struct {
	Offset  int `json:"offset,omitempty"`
	Timeout int `json:"timeout,omitempty"`
}

type User struct {
	ID           int64   `json:"id"`
	IsBot        bool    `json:"is_bot"`
	FirstName    string  `json:"first_name"`
	LastName     *string `json:"last_name,omitempty"`
	Username     *string `json:"username,omitempty"`
	LanguageCode *string `json:"language_code,omitempty"`
	IsPremium    *bool   `json:"is_premium,omitempty"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
}

type CallbackQuery struct {
	ID              string   `json:"id"`
	From            User     `json:"from"`
	Message         *Message `json:"message,omitempty"`
	InlineMessageID string   `json:"inline_message_id,omitempty"`
	ChatInstance    string   `json:"chat_instance"`
	Data            string   `json:"data"`
}
