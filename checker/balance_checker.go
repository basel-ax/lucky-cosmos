package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// cosmosAPI is the base URL for the Cosmos public REST API to fetch account details.
	cosmosAPI = "https://rest.cosmos.directory/cosmoshub/cosmos/bank/v1beta1/balances/%s"
	// atomDenom is the denomination for the native Cosmos Hub coin (micro-atom).
	atomDenom = "uatom"
)

// Balance represents a single coin balance in the API response.
type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

// BalanceResponse represents the structure of the account data from the Cosmos API.
type BalanceResponse struct {
	Balances []Balance `json:"balances"`
}

// CheckBalance checks a Cosmos address's balance using a public Cosmos REST API.
// It returns the balance of 'uatom' as a string. If no 'uatom' balance is found, it returns "0".
func CheckBalance(address string) (string, error) {
	if address == "" {
		return "0", fmt.Errorf("address cannot be empty")
	}

	// Construct the full API URL for the given address.
	url := fmt.Sprintf(cosmosAPI, address)

	// Create a new HTTP client with a reasonable timeout to avoid hanging indefinitely.
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Perform the GET request to the Cosmos API.
	resp, err := client.Get(url)
	if err != nil {
		return "0", fmt.Errorf("failed to make HTTP request to Cosmos API: %w", err)
	}
	defer resp.Body.Close()

	// The API should return a 200 OK status code on success.
	if resp.StatusCode != http.StatusOK {
		// A 404 might mean the account is new and has no transactions, which is not an error for our case.
		if resp.StatusCode == http.StatusNotFound {
			return "0", nil
		}
		return "0", fmt.Errorf("received non-200 status code from Cosmos API: %d", resp.StatusCode)
	}

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "0", fmt.Errorf("failed to read response body: %w", err)
	}

	// Unmarshal the JSON response into our BalanceResponse struct.
	var balanceResp BalanceResponse
	if err := json.Unmarshal(body, &balanceResp); err != nil {
		return "0", fmt.Errorf("failed to unmarshal json response from Cosmos API: %w", err)
	}

	// Iterate through the balances to find the 'uatom' denomination.
	for _, balance := range balanceResp.Balances {
		if balance.Denom == atomDenom {
			// If the amount is empty, treat it as zero.
			if balance.Amount == "" {
				return "0", nil
			}
			return balance.Amount, nil
		}
	}

	// If no 'uatom' balance was found in the balances array.
	return "0", nil
}
