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

### Installing gaiad for Cron Jobs

For the migrate-addresses command to work in cron jobs, `gaiad` must be available in the system PATH.

**Installation**: Download the latest release from https://github.com/cosmos/gaia/releases and extract the binary to your system:

```bash
# Download and extract (recommended v27.2.0)
wget https://github.com/cosmos/gaia/releases/download/v27.2.0/gaia-v27.2.0-linux-amd64.tar.gz
tar -xzf gaia-v27.2.0-linux-amd64.tar.gz
sudo cp gaiad /usr/local/bin/
```

After installation, verify it's available:
```bash
which gaiad
```

If you can't install system-wide, add it to PATH in your cron job:

```bash
# Example cron job with PATH to gaiad
0 * * * * cd /mnt/usb/projects/lucky-cosmos/checker && PATH=$PATH:/home/swenro11/go/bin go run . -migrate-addresses --prod >> migrate-addresses.log 2>&1
```

**Note**: Cron jobs have minimal PATH. If `gaiad` is not in `/usr/local/bin`, you must add its location to PATH in the cron job (e.g., `PATH=$PATH:/home/swenro11/go/bin`).  

## Checker Command

This project includes a standalone command-line tool in the `checker` directory that performs the following actions:

1.  Connects to the PostgreSQL database.
2.  Fetches all `WalletBalance` records.
3.  For each record, if the `CosmosAddress` is missing, it generates one from the mnemonic and saves it to the database.
4.  It then checks the wallet's balance using the public Atomscan API.
5. If the balance is greater than zero and a notification has not been sent yet, it sends a message to a specified Telegram chat with a link to the wallet on Atomscan.
6. Finally, it marks the wallet as notified in the database to prevent duplicate messages.

#### Telegram Topics

If your Telegram bot is in a supergroup with topics enabled, you can send notifications to a specific topic instead of the general chat. To find the message thread ID:
- Open the topic in Telegram
- Right-click on the topic name → "Copy link"
- The link will contain the thread ID (e.g., `https://t.me/c/1234567890/1` where `1` is the thread ID)
- Set this ID in your `.env` file: `TELEGRAM_TOPIC_ID=1`

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

You can limit the number of addresses to migrate using the `-limit` flag or the `MIGRATE_LIMIT` environment variable:

```bash
# Limit to 100 addresses using flag
cd checker
go run . -migrate-addresses --limit=100

# Limit to 50 addresses using environment variable
cd checker
MIGRATE_LIMIT=50 go run . -migrate-addresses
```

If neither is specified, the default limit is 1000 addresses. To disable the limit, set `-limit=0` or `MIGRATE_LIMIT=0`.

In production mode, it also sends a Telegram notification with the number of addresses migrated.

#### Cron Job for Address Migration

To run address migration periodically (e.g., every hour), add this to your crontab:

```bash
0 * * * * cd /path/to/lucky-cosmos/checker && /usr/local/go/bin/go run . -migrate-addresses --prod --limit=100 >> /dev/null 2>&1
```

Or if you have a built binary:

```bash
0 * * * * /path/to/lucky-cosmos-checker -migrate-addresses --prod --limit=100 >> /dev/null 2>&1
```

You can also use the `MIGRATE_LIMIT` environment variable instead of the `-limit` flag:

```bash
0 * * * * MIGRATE_LIMIT=100 /path/to/lucky-cosmos-checker -migrate-addresses --prod >> /dev/null 2>&1
```

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
| `TELEGRAM_TOPIC_ID` | Optional: The message thread ID for sending notifications to a specific topic in a Telegram supergroup forum. |
| `MIGRATE_LIMIT` | Optional: Limit the number of addresses to generate during migration (default: 1000, 0 for no limit). |

## Cron Job Setup

To run the checker command automatically on a schedule, you can set up a cron job. The checker supports a `--prod` flag for production mode.

### Production Mode (`--prod`)

When running with the `--prod` flag:
- Console output is suppressed (logs are written to `/tmp/lucky-cosmos-checker.log`)
- A summary message is sent to Telegram when the command completes, showing how many rows were processed/updated
- A lock file (`/tmp/lucky-cosmos-checker.lock`) prevents multiple instances from running simultaneously

### Example Crontab Entry

To run the checker every 5 minutes in production mode, add the following to your crontab:

```bash
*/5 * * * * cd /path/to/lucky-cosmos/checker && /usr/local/go/bin/go run . --prod >> /dev/null 2>&1
```

Or if you have a built binary:

```bash
*/5 * * * * /path/to/lucky-cosmos-checker >> /dev/null 2>&1
```

To edit your crontab, run:

```bash
crontab -e
```

### Monitoring

- Check the log file for production mode: `tail -f /tmp/lucky-cosmos-checker.log`
- The lock file prevents concurrent runs - if you see "Another instance is already running" in the logs, a previous cron job may still be processing
