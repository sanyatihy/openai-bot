package telegram

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/sanyatihy/openai-bot/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSendMessage(t *testing.T) {
	tests := []struct {
		name           string
		requestOptions *SendMessageRequest
		mockResponse   *http.Response
		mockError      error
		expectedResult *Message
		expectedError  error
	}{
		{
			name: "Success",
			requestOptions: &SendMessageRequest{
				ChatID: 12345,
				Text:   "U here?",
			},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"ok": true,
					"result": {
						"message_id": 1,
						"text": "U here?",
						"chat": {
							"id": 12345
						}
					}
				}`))),
			},
			expectedResult: &Message{
				MessageID: 1,
				Text:      utils.StringPtr("U here?"),
				Chat: Chat{
					ID: 12345,
				},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "Error",
			requestOptions: &SendMessageRequest{
				ChatID: 12345,
				Text:   "U here?",
			},
			mockResponse:   nil,
			mockError:      errors.New("err"),
			expectedResult: nil,
			expectedError: &InternalError{
				Message: fmt.Sprintf("error making request: %s", errors.New("err")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHTTPClient := new(MockHTTPClient)

			mockHTTPClient.On("Do", mock.Anything).Return(tt.mockResponse, tt.mockError)

			mockClient := NewBotClient(mockHTTPClient, "test_token")

			response, err := mockClient.SendMessage(context.Background(), tt.requestOptions)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, response)

			mockHTTPClient.AssertCalled(t, "Do", mock.Anything)
		})
	}
}
