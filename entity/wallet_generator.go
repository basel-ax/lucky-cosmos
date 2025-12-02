package entity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const (
	// commandName is the name of the Cosmos SDK CLI tool.
	// This should be in the system's PATH. It's assumed to be `gaiad`
	// or another compatible binary.
	commandName = "gaiad"

	// keyringBackend is the backend to use for the temporary key.
	// "test" is an insecure backend that doesn't require a password
	// and is suitable for this temporary use case.
	keyringBackend = "test"
)

// keyOutput represents the JSON output from the `gaiad keys add` command.
type keyOutput struct {
	Address string `json:"address"`
}

// SetCosmosAddressFromMnemonic derives a Cosmos address from the Mnemonic field
// and sets the CosmosAddress field. It uses an external CLI tool to avoid Go-level
// dependencies on the Cosmos SDK.
// This assumes a Cosmos SDK-based command-line tool (e.g., gaiad) is installed
// and available in the system's PATH.
func (wb *WalletBalance) SetCosmosAddressFromMnemonic() error {
	if strings.TrimSpace(wb.Mnemonic) == "" {
		return fmt.Errorf("mnemonic is empty")
	}

	// Use a unique name for the key to avoid conflicts if the command is run concurrently.
	// This key will be deleted after use.
	keyName := fmt.Sprintf("tempkey-%d", time.Now().UnixNano())

	// The command to add a key from a mnemonic (recover).
	// --keyring-backend test: Uses an insecure, temporary keyring.
	// --output json: Ensures the output is machine-readable.
	cmd := exec.Command(commandName, "keys", "add", keyName, "--recover", "--keyring-backend", keyringBackend, "--output", "json")

	// The `keys add --recover` command expects the mnemonic from stdin.
	cmd.Stdin = strings.NewReader(wb.Mnemonic)

	var outJSON, errJSON bytes.Buffer
	cmd.Stdout = &outJSON
	cmd.Stderr = &errJSON

	// Clean up the key from the test keyring after we are done.
	defer deleteKey(keyName)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute '%s': %s: %w", cmd.String(), errJSON.String(), err)
	}

	var result keyOutput
	if err := json.Unmarshal(outJSON.Bytes(), &result); err != nil {
		return fmt.Errorf("failed to parse command output: %w", err)
	}

	if result.Address == "" {
		return fmt.Errorf("address not found in command output")
	}

	wb.CosmosAddress = result.Address
	return nil
}

// deleteKey removes a key from the test keyring.
// This is a cleanup utility to avoid polluting the keyring with temporary keys.
func deleteKey(keyName string) {
	// The --yes flag is used to automatically confirm the deletion.
	cmd := exec.Command(commandName, "keys", "delete", keyName, "--keyring-backend", keyringBackend, "--yes")
	// We run this on a best-effort basis and ignore potential errors,
	// as the key might not exist if the initial add command failed.
	_ = cmd.Run()
}
