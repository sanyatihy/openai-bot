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

func TestGetUpdates(t *testing.T) {
	tests := []struct {
		name           string
		requestOptions *GetUpdatesRequest
		mockResponse   *http.Response
		mockError      error
		expectedResult []Update
		expectedError  error
	}{
		{
			name: "Success",
			requestOptions: &GetUpdatesRequest{
				Offset:  0,
				Timeout: 0,
			},
			mockResponse: &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewReader([]byte(`{
					"ok": true,
					"result": [
						{
							"update_id": 1,
							"message": {
								"message_id": 1,
								"text": "U here?",
								"chat": {
									"id": 12345
								}
							}
						}
					]
				}`))),
			},
			expectedResult: []Update{
				{
					UpdateID: 1,
					Message: Message{
						MessageID: 1,
						Text:      utils.StringPtr("U here?"),
						Chat: Chat{
							ID: 12345,
						},
					},
				},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "Error",
			requestOptions: &GetUpdatesRequest{
				Offset:  0,
				Timeout: 0,
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

			botClient := NewBotClient(mockHTTPClient, "test_token")

			response, err := botClient.GetUpdates(context.Background(), tt.requestOptions)

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
