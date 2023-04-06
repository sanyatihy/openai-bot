package telegram

import (
	"context"
	"net/http"
)

type BotClient interface {
	GetUpdates(ctx context.Context, requestOptions *GetUpdatesRequest) ([]Update, error)
	SendMessage(ctx context.Context, requestOptions *SendMessageRequest) (*Message, error)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
