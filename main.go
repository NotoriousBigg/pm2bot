package main

import (
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

type Config struct {
	BotToken string
	AdminID  int64
	Debug    bool
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("❌ Error: BOT_TOKEN is not set in environment or .env file")
	}

	adminIDStr := os.Getenv("ADMIN_ID")
	if adminIDStr == "" {
		log.Fatal("❌ Error: ADMIN_ID is not set. You must specify your Telegram ID.")
	}

	adminID, err := strconv.ParseInt(adminIDStr, 10, 64)
	if err != nil {
		log.Fatal("❌ Error: ADMIN_ID must be a valid integer")
	}

	debug := false
	if os.Getenv("DEBUG") == "true" {
		debug = true
	}

	return &Config{
		BotToken: token,
		AdminID:  adminID,
		Debug:    debug,
	}
}

func main() {
	config := LoadConfig()
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		log.Panic("Failed to create bot: ", err)
	}

	bot.Debug = config.Debug
	log.Printf("🤖 Authorized on account %s", bot.Self.UserName)
	log.Printf("Twisted Manager is active. Waiting for commands from Admin ID: %d", config.AdminID)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		go HandleUpdate(bot, update, config.AdminID)
	}
}
