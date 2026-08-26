package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gamee1910/volt/config"
	"github.com/gamee1910/volt/internal/infrastructure/di"
	"github.com/gamee1910/volt/pkg/logger"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()
	log := logger.NewLogger(cfg.ApplicationConfig.Env)

	databaseConnection, err := config.DatabaseConnection(cfg)
	if err != nil {
		log.Fatalf("[report] database error: ", "error", err)
	}
	defer closeDB(databaseConnection, log)

	container, err := di.NewContainer(cfg, databaseConnection, log)
	if err != nil {
		log.Fatalf("[report] container error: ", "error", err)
	}

	chatID, err := fetchFirstChatID(cfg.ApplicationConfig.TelegramConfig.TelegramAPIKey)
	if err != nil {
		log.Fatalf("[report] failed to get chat ID", "error", err)
	}

	if err := container.ElectricityCommand().HandleYesterdayCommand(ctx, chatID); err != nil {
		log.Error("[report] failed to send yesterday report", "error", err)
	}
}

func fetchFirstChatID(apiKey string) (int64, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?limit=1", apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("getUpdates request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result []struct {
			Message struct {
				Chat struct {
					ID int64 `json:"id"`
				} `json:"chat"`
			} `json:"message"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode getUpdates response: %w", err)
	}
	if !result.OK || len(result.Result) == 0 {
		return 0, fmt.Errorf("no updates found, make sure the bot has received at least one message")
	}
	return result.Result[0].Message.Chat.ID, nil
}

func closeDB(db *sql.DB, log *logger.Logger) {
	if err := db.Close(); err != nil {
		log.Error("[report] failed to close database", "error", err)
	}
}
