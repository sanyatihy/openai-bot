package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *botClient) doRequest(ctx context.Context, method, endpoint string, requestData interface{}) (*http.Response, error) {
	var reqBody bytes.Buffer

	if requestData != nil {
		encoder := json.NewEncoder(&reqBody)
		if err := encoder.Encode(requestData); err != nil {
			return nil, fmt.Errorf("error encoding request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, &reqBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	c.setDefaultHeaders(req)

	return c.httpClient.Do(req)
}

func (c *botClient) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

func (c *botClient) processResponseBody(resp *http.Response, target interface{}) error {
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("error decoding response body: %w", err)
	}

	return nil
}

func (c *botClient) checkStatusCode(resp *http.Response) error {
	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusTooManyRequests:
		return fmt.Errorf("too many requests: %d", resp.StatusCode)
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
		return fmt.Errorf("client error: %d", resp.StatusCode)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return fmt.Errorf("server error: %d", resp.StatusCode)
	default:
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
