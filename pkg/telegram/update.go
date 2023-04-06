package telegram

import (
	"context"
	"fmt"
	"net/http"
)

func (c *botClient) GetUpdates(ctx context.Context, requestOptions *GetUpdatesRequest) ([]Update, error) {
	url := fmt.Sprintf("%s%s/getUpdates", baseURL, c.token)

	resp, err := c.doRequest(ctx, http.MethodPost, url, requestOptions)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if err := c.checkStatusCode(resp); err != nil {
		return nil, err
	}

	var response struct {
		OK      bool     `json:"ok"`
		Updates []Update `json:"result"`
		Error   APIError `json:"error"`
	}
	if err := c.processResponseBody(resp, &response); err != nil {
		return nil, err
	}

	if !response.OK {
		return nil, fmt.Errorf("error: %s", response.Error.Description)
	}

	return response.Updates, nil
}
