package main

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// sendTelegramNotification sends a message to a specified Telegram chat.
// It notifies the user about a wallet with a positive balance or a custom message.
func sendTelegramNotification(botToken, chatIDStr, messageText string) error {
	// Convert the chat ID string from environment variables to an int64,
	// which is required by the Telegram Bot API library.
	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TELEGRAM_CHAT_ID, must be an integer: %w", err)
	}

	// Get message thread ID for topics/forums (optional)
	messageThreadID := getMessageThreadID()

	// Initialize a new bot instance with the provided token.
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Check if the message is a custom summary message (starts with ✅)
	if strings.HasPrefix(messageText, "✅") {
		// Custom summary message - don't add the Atomscan link
		// messageText already contains the full message
	} else {
		// Wallet address message - append Atomscan link
		messageText = fmt.Sprintf(
			"🎉 Lucky wallet found! 🎉\n\nAddress: %s\n\nBalance is greater than 0.\n\nView on Atomscan:\nhttps://www.atomscan.com/accounts/%s",
			messageText,
			messageText,
		)
	}

	// Create a new message configuration.
	msg := tgbotapi.NewMessage(chatID, messageText)

	if messageThreadID != 0 {
		v := reflect.ValueOf(&msg)
		if v.Elem().FieldByName("MessageThreadID").IsValid() {
			v.Elem().FieldByName("MessageThreadID").SetInt(messageThreadID)
		}
	}

	// Send the message.
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}

	log.Printf("Successfully sent notification: %s", messageText)
	return nil
}

func getMessageThreadID() int64 {
	threadIDStr := os.Getenv("TELEGRAM_MESSAGE_THREAD_ID")
	if threadIDStr == "" {
		return 0
	}

	threadID, err := strconv.ParseInt(threadIDStr, 10, 64)
	if err != nil {
		log.Printf("Warning: Invalid TELEGRAM_MESSAGE_THREAD_ID '%s': %v", threadIDStr, err)
		return 0
	}
	return threadID
}
