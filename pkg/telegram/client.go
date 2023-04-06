package telegram

const (
	baseURL = "https://api.telegram.org/bot"
)

type botClient struct {
	httpClient httpClient
	token      string
}

func NewBotClient(httpClient httpClient, token string) BotClient {
	return &botClient{
		httpClient: httpClient,
		token:      token,
	}
}
