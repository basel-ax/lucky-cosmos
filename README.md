# lucky-cosmos

This project provides a Go library to generate a Cosmos blockchain wallet address from a 12-word mnemonic phrase. It is designed to be used in applications that need to interact with the Cosmos blockchain for wallet creation and balance checking. The library also includes functionality to store wallet information in a PostgreSQL database using GORM.

## Features

-   Generate a Cosmos wallet address from a 12-word BIP39 mnemonic.
-   Uses an external CLI tool (`gaiad`) for address generation, removing complex Go dependencies.
-   Database integration with GORM for storing wallet data.
-   Designed to be configured via environment variables.

## Usage

The primary entity in this library is `WalletBalance`. To generate a Cosmos address, you can create an instance of `WalletBalance`, set the `Mnemonic` field, and then call the `SetCosmosAddressFromMnemonic` method.

```go
package main

import (
	"fmt"
	"log"

	"github.com/basel-ax/lucky-cosmos/entity"
)

func main() {
	// Example mnemonic. Replace with your actual mnemonic.
	mnemonic := "guard cream sadness conduct invite crumble farm vendor index song man myth"

	wallet := &entity.WalletBalance{
		Mnemonic: mnemonic,
	}

	err := wallet.SetCosmosAddressFromMnemonic()
	if err != nil {
		log.Fatalf("Failed to generate Cosmos address: %v", err)
	}

	fmt.Printf("Generated Cosmos Address: %s\n", wallet.CosmosAddress)

	// Here you would typically save the 'wallet' object to your database.
}
```

**Note:** This usage example assumes you have the `gaiad` command-line tool (or a compatible Cosmos SDK binary) installed and available in your system's `PATH`.

## Configuration

This library is configured using environment variables. You can create a `.env` file in your project root or set the variables in your deployment environment.

| Variable | Description |
|---|---|
| `DB_HOST` | The hostname of the PostgreSQL database. |
| `DB_PORT` | The port of the PostgreSQL database. |
| `DB_USER` | The username for the PostgreSQL database. |
| `DB_PASSWORD` | The password for the PostgreSQL database. |
| `DB_NAME` | The name of the PostgreSQL database. |
| `TELEGRAM_BOT_TOKEN` | The token for your Telegram bot. |
| `TELEGRAM_CHAT_ID` | The chat ID to send notifications to. |
