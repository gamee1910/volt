package telegram

import "net/http"

type TelegramClient struct {
	apiKey string
}

func (t *TelegramClient) SendMessage(w http.ResponseWriter, r *http.Request) {
	return
}
