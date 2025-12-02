package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/basel-ax/lucky-cosmos/entity"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load .env file for local development from the parent directory.
	// In production, environment variables should be set directly.
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, relying on system environment variables")
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	telegramToken := os.Getenv("TELEGRAM_APP_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	if dbHost == "" || dbPort == "" || dbUser == "" || dbPassword == "" || dbName == "" {
		log.Fatal("Database environment variables (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME) must be set")
	}

	if telegramToken == "" || telegramChatID == "" {
		log.Fatal("Telegram environment variables (TELEGRAM_APP_BOT_TOKEN, TELEGRAM_CHAT_ID) must be set")
	}

	// Setup Database Connection
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName, dbPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate the schema to ensure the table exists and is up-to-date.
	if err := db.AutoMigrate(&entity.WalletBalance{}); err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}

	log.Println("Checker application starting...")

	// Fetch all wallet balances from the database that have a Cosmos address.
	var wallets []entity.WalletBalance
	if err := db.Where("cosmos_address != ?", "").Find(&wallets).Error; err != nil {
		log.Fatalf("Error fetching wallets from database: %v", err)
	}

	log.Printf("Found %d wallets to process.\n", len(wallets))

	for _, wallet := range wallets {
		log.Printf("Processing wallet ID: %d", wallet.ID)
		currentWallet := wallet // Make a mutable copy

		// 2. Check wallet balance if not already notified.
		if !currentWallet.IsNotified {
			log.Printf("Checking balance for address: %s", currentWallet.CosmosAddress)
			hasBalance, err := CheckBalance(currentWallet.CosmosAddress)
			if err != nil {
				log.Printf("ERROR: Could not check balance for address %s (ID: %d): %v", currentWallet.CosmosAddress, currentWallet.ID, err)
				continue // Skip to next wallet.
			}

			now := time.Now()
			currentWallet.BalanceUpdatedAt = &now

			// 3. If balance > 0, send a Telegram message.
			if hasBalance {
				log.Printf("SUCCESS: Wallet %s (ID: %d) has a positive balance!", currentWallet.CosmosAddress, currentWallet.ID)

				if err := sendTelegramNotification(telegramToken, telegramChatID, currentWallet.CosmosAddress); err != nil {
					log.Printf("ERROR: Failed to send Telegram notification for wallet %s (ID: %d): %v", currentWallet.CosmosAddress, currentWallet.ID, err)
				} else {
					// 4. Mark as notified to prevent spam.
					currentWallet.IsNotified = true
				}
			} else {
				log.Printf("Wallet %s (ID: %d) has zero balance.", currentWallet.CosmosAddress, currentWallet.ID)
			}

			// Save the updated wallet state (BalanceUpdatedAt and possibly IsNotified)
			if err := db.Save(&currentWallet).Error; err != nil {
				log.Printf("ERROR: Failed to update wallet state for wallet %s (ID: %d): %v", currentWallet.CosmosAddress, currentWallet.ID, err)
			} else {
				if hasBalance && currentWallet.IsNotified {
					log.Printf("Successfully marked wallet %s (ID: %d) as notified.", currentWallet.CosmosAddress, currentWallet.ID)
				} else {
					log.Printf("Successfully updated balance timestamp for wallet %s (ID: %d).", currentWallet.CosmosAddress, currentWallet.ID)
				}
			}
		} else {
			log.Printf("Wallet %s (ID: %d) was already notified. Skipping.", currentWallet.CosmosAddress, currentWallet.ID)
		}
	}

	log.Println("Checker application finished.")
}
