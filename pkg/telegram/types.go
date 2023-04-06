package telegram

type Update struct {
	UpdateID int     `json:"update_id"`
	Message  Message `json:"message"`
}

type Message struct {
	MessageID int    `json:"message_id"`
	Text      string `json:"text"`
	Chat      Chat   `json:"chat"`
}

type Chat struct {
	ID int `json:"id"`
}

type SendMessageRequest struct {
	ChatID int    `json:"chat_id"`
	Text   string `json:"text"`
}

type APIError struct {
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type GetUpdatesRequest struct {
	Offset int `json:"offset,omitempty"`
}
