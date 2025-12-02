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
	// Load .env file for local development.
	// In production, environment variables should be set directly.
	if err := godotenv.Load(); err != nil {
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

	// Fetch all wallet balances from the database.
	var wallets []entity.WalletBalance
	if err := db.Find(&wallets).Error; err != nil {
		log.Fatalf("Error fetching wallets from database: %v", err)
	}

	log.Printf("Found %d wallets to process.\n", len(wallets))

	for _, wallet := range wallets {
		log.Printf("Processing wallet ID: %d", wallet.ID)
		currentWallet := wallet // Make a mutable copy

		// 1. Check if CosmosAddress is empty. If so, generate it and save to DB.
		if currentWallet.CosmosAddress == "" {
			log.Printf("Wallet ID %d has no address, generating...", currentWallet.ID)
			if err := currentWallet.SetCosmosAddressFromMnemonic(); err != nil {
				log.Printf("ERROR: Could not generate address for wallet ID %d: %v", currentWallet.ID, err)
				continue // Skip to the next wallet.
			}

			if err := db.Save(&currentWallet).Error; err != nil {
				log.Printf("ERROR: Could not save newly generated address for wallet ID %d: %v", currentWallet.ID, err)
				continue // Skip to the next wallet.
			}
			log.Printf("Successfully generated and saved address for wallet ID %d: %s", currentWallet.ID, currentWallet.CosmosAddress)
		}

		// 2. Check wallet balance if not already notified.
		if !currentWallet.IsNotified {
			log.Printf("Checking balance for address: %s", currentWallet.CosmosAddress)
			hasBalance, err := CheckBalance(currentWallet.CosmosAddress)
			if err != nil {
				log.Printf("ERROR: Could not check balance for address %s (ID: %d): %v", currentWallet.CosmosAddress, currentWallet.ID, err)
				continue // Skip to next wallet.
			}

			// 3. If balance > 0, send a Telegram message.
			if hasBalance {
				log.Printf("SUCCESS: Wallet %s (ID: %d) has a positive balance!", currentWallet.CosmosAddress, currentWallet.ID)

				if err := sendTelegramNotification(telegramToken, telegramChatID, currentWallet.CosmosAddress); err != nil {
					log.Printf("ERROR: Failed to send Telegram notification for wallet %s (ID: %d): %v", currentWallet.CosmosAddress, currentWallet.ID, err)
					continue // Don't mark as notified if the notification failed.
				}

				// 4. Mark as notified in the DB to prevent spam.
				now := time.Now()
				currentWallet.IsNotified = true
				currentWallet.BalanceUpdatedAt = &now
				if err := db.Save(&currentWallet).Error; err != nil {
					log.Printf("ERROR: Failed to update IsNotified flag for wallet %s (ID: %d): %v", currentWallet.CosmosAddress, currentWallet.ID, err)
				} else {
					log.Printf("Successfully marked wallet %s (ID: %d) as notified.", currentWallet.CosmosAddress, currentWallet.ID)
				}
			} else {
				log.Printf("Wallet %s (ID: %d) has zero balance.", currentWallet.CosmosAddress, currentWallet.ID)
			}
		} else {
			log.Printf("Wallet %s (ID: %d) was already notified. Skipping.", currentWallet.CosmosAddress, currentWallet.ID)
		}
	}

	log.Println("Checker application finished.")
}
