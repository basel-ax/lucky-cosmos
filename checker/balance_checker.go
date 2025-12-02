package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"
)

const (
	// atomScanAPIBase is the base URL for the Atomscan API to fetch account details.
	atomScanAPIBase = "https://api.atomscan.com/accounts/%s"
	// atomDenom is the denomination for the native Cosmos Hub coin (micro-atom).
	atomDenom = "uatom"
)

// Balance represents a single coin balance in the API response.
type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// AccountInfo represents the structure of the account data from the Atomscan API.
// This is a simplified version focusing only on the balances we need.
type AccountInfo struct {
	Balances []Balance `json:"balances"`
}

// CheckBalance checks a Cosmos address's balance using the public Atomscan API.
// It returns true if the balance of 'uatom' is greater than zero, and false otherwise.
func CheckBalance(address string) (bool, error) {
	if address == "" {
		return false, fmt.Errorf("address cannot be empty")
	}

	// Construct the full API URL for the given address.
	url := fmt.Sprintf(atomScanAPIBase, address)

	// Create a new HTTP client with a reasonable timeout to avoid hanging indefinitely.
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Perform the GET request to the Atomscan API.
	resp, err := client.Get(url)
	if err != nil {
		return false, fmt.Errorf("failed to make HTTP request to atomscan: %w", err)
	}
	defer resp.Body.Close()

	// The API should return a 200 OK status code on success.
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("received non-200 status code from atomscan: %d", resp.StatusCode)
	}

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON response into our AccountInfo struct.
	var accInfo AccountInfo
	if err := json.Unmarshal(body, &accInfo); err != nil {
		return false, fmt.Errorf("failed to unmarshal json response from atomscan: %w", err)
	}

	// Iterate through the balances to find the 'uatom' denomination.
	for _, balance := range accInfo.Balances {
		if balance.Denom == atomDenom {
			// Convert the amount string to a big.Int for safe comparison of large numbers.
			balanceAmount, ok := new(big.Int).SetString(balance.Amount, 10)
			if !ok {
				// This case would happen if the amount is not a valid integer string.
				return false, fmt.Errorf("failed to parse balance amount string: '%s'", balance.Amount)
			}

			// Check if the balance is greater than zero.
			if balanceAmount.Cmp(big.NewInt(0)) > 0 {
				return true, nil
			}
		}
	}

	// If no 'uatom' balance was found or its value was 0.
	return false, nil
}
