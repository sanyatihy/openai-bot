package telegram

import (
	"context"
	"fmt"
	"net/http"
)

func (c *botClient) SendMessage(ctx context.Context, requestOptions *SendMessageRequest) (*Message, error) {
	url := fmt.Sprintf("%s%s/sendMessage", baseURL, c.token)

	resp, err := c.doRequest(ctx, http.MethodPost, url, requestOptions)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var response struct {
		OK      bool     `json:"ok"`
		Message Message  `json:"result"`
		Error   APIError `json:"error"`
	}
	if err := c.processResponseBody(resp, &response); err != nil {
		return nil, err
	}

	if !response.OK {
		return nil, &response.Error
	}

	return &response.Message, nil
}
