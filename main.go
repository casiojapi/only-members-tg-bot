package main

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Get the bot token from the environment variable or set it here
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	// Create a new Bot instance
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	// Enable debugging (optional)
	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Create a new UpdateConfig with a timeout of 60 seconds
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Start receiving updates
	updates := bot.GetUpdatesChan(u)

    for update := range updates {
        if update.Message == nil { // ignore any non-Message updates
            continue
        }

        if !update.Message.IsCommand() { // ignore any non-command Messages
            continue
        }

        // Create a new MessageConfig. We don't have text yet,
        // so we leave it empty.
        msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

        // Extract the command from the Message.
        switch update.Message.Command() {
        case "proof":
		msg.Text = "proof of membership to the group: \n 0xbc20bdb20b219d31ad12b219422f9132d21a92"
        case "parsley":
            msg.Text = "pollera"
        case "vitalik":
            msg.Text = "marciano"
        default:
            msg.Text = "I don't know that command"
        }

        if _, err := bot.Send(msg); err != nil {
            log.Panic(err)
        }
    }
}
