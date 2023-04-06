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
			return nil, &InternalError{
				Message: fmt.Sprintf("error encoding request body: %s", err),
			}
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, &reqBody)
	if err != nil {
		return nil, &InternalError{
			Message: fmt.Sprintf("error creating request: %s", err),
		}
	}

	c.setDefaultHeaders(req)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, &InternalError{
			Message: fmt.Sprintf("error making request: %s", err),
		}
	}

	return res, nil
}

func (c *botClient) setDefaultHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
}

func (c *botClient) processResponseBody(resp *http.Response, target interface{}) error {
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(target); err != nil {
		return &InternalError{
			Message: fmt.Sprintf("error decoding response body: %s", err),
		}
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
