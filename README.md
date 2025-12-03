# lucky-cosmos

This project provides a Go library to generate a Cosmos blockchain wallet address from a 12-word mnemonic phrase. It is designed to be used in applications that need to interact with the Cosmos blockchain for wallet creation and balance checking. The library also includes functionality to store wallet information in a PostgreSQL database using GORM.

## LuckySix
This is a part of the project. All results you can get by 3 projects.    
 - [LuckySix](https://github.com/basel-ax/luckysix)
 - [LuckyEth](https://github.com/basel-ax/lucky-eth)
 - [LuckyCosmos](https://github.com/basel-ax/lucky-cosmos)

## Features

-   Generate a Cosmos wallet address from a 12-word BIP39 mnemonic.
-   Uses an external CLI tool (`gaiad`) for address generation, removing complex Go dependencies.
-   Database integration with GORM for storing wallet data.
-   A standalone checker command to monitor wallet balances and send notifications.
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

[Installing Gaia](https://docs.cosmos.network/hub/v25/getting-started/installation)  

## Checker Command

This project includes a standalone command-line tool in the `checker` directory that performs the following actions:

1.  Connects to the PostgreSQL database.
2.  Fetches all `WalletBalance` records.
3.  For each record, if the `CosmosAddress` is missing, it generates one from the mnemonic and saves it to the database.
4.  It then checks the wallet's balance using the public Atomscan API.
5.  If the balance is greater than zero and a notification has not been sent yet, it sends a message to a specified Telegram chat with a link to the wallet on Atomscan.
6.  Finally, it marks the wallet as notified in the database to prevent duplicate messages.

### Running the Checker

To run the checker, navigate to the `checker` directory and run the `main.go` file:

```bash
cd checker
go run .
```

Make sure you have all the required environment variables set before running the command.

### Migrating Addresses

If you have existing wallets in your database that have a `Mnemonic` but are missing a `CosmosAddress`, you can use the `-migrate-addresses` flag to generate and save the addresses for them.

```bash
cd checker
go run . -migrate-addresses
```

This command will find all wallets that need an address, generate it, and save it to the database.

## Configuration

This library is configured using environment variables. You can create a `.env` file in your project root or set the variables in your deployment environment.

| Variable | Description |
|---|---|
| `DB_HOST` | The hostname of the PostgreSQL database. |
| `DB_PORT` | The port of the PostgreSQL database. |
| `DB_USER` | The username for the PostgreSQL database. |
| `DB_PASSWORD` | The password for the PostgreSQL database. |
| `DB_NAME` | The name of the PostgreSQL database. |
| `TELEGRAM_APP_BOT_TOKEN` | The token for your Telegram bot, used by the checker command. |
| `TELEGRAM_CHAT_ID` | The chat ID to send notifications to. |
