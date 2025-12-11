package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// ========== TESTING CONFIGURATION (TESTNET ONLY) ==========
	// NOTE: Currently configured for Shasta Testnet
	// For testing on Shasta Testnet (free TRX from faucet https://www.tronweb.org/shasta)
	// Using native TRX instead of USDT for simplicity and faster testing
	TRON_RPC_URL      = "https://api.shasta.trongrid.io" // Testnet API endpoint
	USE_TRX_FOR_TEST  = true                              // Set to false for USDT on mainnet - CURRENTLY IN TEST MODE

	// ========== PRODUCTION CONFIGURATION (CHANGE FOR MAINNET) ==========
	// TODO: Switch to mainnet for production - uncomment lines below
	// TRON_RPC_URL      = "https://api.trongrid.io"
	// USE_TRX_FOR_TEST  = false
	
	// TODO: Replace with real USDT contract address on mainnet
	USDT_CONTRACT     = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t" // USDT on Tron mainnet
	
	USDT_DECIMALS     = 6 // Decimal places for USDT token
	TRX_DECIMALS      = 6 // Decimal places for native TRX
	CONFIRMATION_NUM  = 25 // Number of blocks to wait for transaction confirmation
)

// TronService provides methods to interact with the Tron blockchain
// Currently configured for Shasta Testnet - switch to mainnet for production
type TronService struct {
	apiKey      string // TronGrid API key for authenticated requests
	mainAddress string // Main wallet address for receiving payments
}

// TronBalance represents the balance information for a Tron address
type TronBalance struct {
	Amount    int64  `json:"amount"`    // Balance amount in smallest units (sun/wei)
	Decimals  int    `json:"decimals"`  // Decimal places (6 for TRX and USDT)
	Address   string `json:"address"`   // The Tron address being queried
	Timestamp int64  `json:"timestamp"` // Query timestamp
}

// TronTransaction represents a Tron blockchain transaction
type TronTransaction struct {
	TxID          string `json:"txID"`            // Transaction ID (hash)
	From          string `json:"from"`            // Sender address
	To            string `json:"to"`              // Recipient address
	Amount        int64  `json:"amount"`          // Transaction amount in smallest units
	ContractAddr  string `json:"contractAddress"` // Smart contract address (for token transfers)
	BlockNumber   int64  `json:"blockNumber"`     // Block number containing the transaction
	Confirmed     bool   `json:"confirmed"`       // Whether transaction has sufficient confirmations
	Timestamp     int64  `json:"timestamp"`       // Transaction timestamp
}

// TronRPCResponse represents a response from the Tron RPC API
type TronRPCResponse struct {
	Result      json.RawMessage `json:"result"`  // Raw JSON result data
	TxID        string          `json:"txID"`    // Transaction ID from response
	Error       string          `json:"error"`   // Error message if request failed
	Code        string          `json:"code"`    // Error code
	Message     string          `json:"message"` // Human-readable message
}

// TronAccountInfo represents basic account information from Tron
type TronAccountInfo struct {
	Address string `json:"address"` // Account address
	Balance int64  `json:"balance"` // Account balance in sun
}

func NewTronService(apiKey, mainAddress string) *TronService {
	return &TronService{
		apiKey:      apiKey,
		mainAddress: mainAddress,
	}
}

// GetMainAddress returns the main Tron address for receiving payments
func (s *TronService) GetMainAddress() string {
	return s.mainAddress
}

// CheckBalance checks TRX or USDT balance on a Tron address
func (s *TronService) CheckBalance(address string) (*TronBalance, error) {
	if USE_TRX_FOR_TEST {
		return s.checkTRXBalance(address)
	}
	return s.checkUSDTBalance(address)
}

// checkTRXBalance checks native TRX balance for testing on Shasta Testnet
// TESTNET MODE: This method is used because we're testing with native TRX instead of USDT
// NOTE: This is for Shasta Testnet only. For mainnet USDT, use checkUSDTBalance()
func (s *TronService) checkTRXBalance(address string) (*TronBalance, error) {
	// Use v1 REST API endpoint (works better with Base58 addresses than /walletsolidity endpoints)
	// NOTE: Debug logs included - REMOVE fmt.Printf statements in production
	url := TRON_RPC_URL + "/v1/accounts/" + address
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if s.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", s.apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TESTNET DEBUG: Print raw API response (remove for production)
	fmt.Printf("[DEBUG] Raw response for %s: %s\n", address, string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// TESTNET DEBUG: Print parsed JSON keys (remove for production)
	fmt.Printf("[DEBUG] Parsed JSON keys: %v\n", getKeys(result))

	var amount int64 = 0
	
	// Parse balance from API response structure
	// Try to get balance from data array
	if data, ok := result["data"].([]interface{}); ok && len(data) > 0 {
		if accountData, ok := data[0].(map[string]interface{}); ok {
			if balance, ok := accountData["balance"].(float64); ok {
				amount = int64(balance)
				// TESTNET DEBUG: Log found balance (remove for production)
				fmt.Printf("[DEBUG] Found balance: %d\n", amount)
			}
		}
	}

	return &TronBalance{
		Amount:    amount,
		Decimals:  TRX_DECIMALS,
		Address:   address,
		Timestamp: time.Now().Unix(),
	}, nil
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// checkUSDTBalance checks USDT token balance (for mainnet production)
// NOT USED IN TEST MODE - See checkTRXBalance for testnet balance checking
// NOTE: This method needs proper Base58 to Hex conversion to work correctly
// Currently it has issues with address format conversion - NEEDS TO BE FIXED BEFORE MAINNET SWITCH
func (s *TronService) checkUSDTBalance(address string) (*TronBalance, error) {
	// Convert address from Base58 (starts with 'T') to Hex format required by /walletsolidity endpoints
	// TODO: Implement proper Base58 decode to Hex conversion (currently a stub returning address as-is)
	hexAddr, err := s.addressToHex(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %v", err)
	}

	// Get USDT balance (TRC-20 token transfer history)
	// We check token balance by querying contract
	usdtHex, err := s.addressToHex(USDT_CONTRACT)
	if err != nil {
		return nil, fmt.Errorf("invalid usdt contract: %v", err)
	}

	balancePayload := map[string]interface{}{
		"owner_address":     hexAddr,
		"contract_address":  usdtHex,
		"function_selector": "balanceOf(address)",
		"parameter":         hexAddr,
	}

	balanceResp, err := s.callTronAPI("/walletsolidity/triggerconstantcontract", balancePayload)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(balanceResp, &result); err != nil {
		return nil, err
	}

	// Parse balance from constant_result
	var amount int64 = 0
	if constantResult, ok := result["constant_result"].([]interface{}); ok && len(constantResult) > 0 {
		if resultStr, ok := constantResult[0].(string); ok {
			// Convert hex string to int64
			if val, err := parseHexBalance(resultStr); err == nil {
				amount = val
			}
		}
	}

	return &TronBalance{
		Amount:    amount,
		Decimals:  USDT_DECIMALS,
		Address:   address,
		Timestamp: time.Now().Unix(),
	}, nil
}

// GetTransaction retrieves detailed transaction information from Tron blockchain by transaction ID
func (s *TronService) GetTransaction(txID string) (*TronTransaction, error) {
	payload := map[string]interface{}{
		"value": txID,
	}

	respBody, err := s.callTronAPI("/walletsolidity/gettransactionbyid", payload)
	if err != nil {
		return nil, err
	}

	var txData map[string]interface{}
	if err := json.Unmarshal(respBody, &txData); err != nil {
		return nil, err
	}

	// Parse transaction
	tx := &TronTransaction{
		TxID:      txID,
		Timestamp: time.Now().Unix(),
	}

	if rawTx, ok := txData["raw_data"].(map[string]interface{}); ok {
		if txBody, ok := rawTx["contract"].([]interface{}); ok && len(txBody) > 0 {
			if contract, ok := txBody[0].(map[string]interface{}); ok {
				if parameter, ok := contract["parameter"].(map[string]interface{}); ok {
					if value, ok := parameter["value"].(map[string]interface{}); ok {
						tx.From = toString(value["owner_address"])
						tx.To = toString(value["to_address"])
						if amount, ok := value["amount"].(float64); ok {
							tx.Amount = int64(amount)
						}
					}
				}
			}
		}
	}

	// Check if confirmed (has block_number)
	if blockNum, ok := txData["blockNumber"].(float64); ok {
		tx.BlockNumber = int64(blockNum)
		tx.Confirmed = true
	}

	return tx, nil
}

// GetAddressTransactions retrieves recent TRC20 token transactions for an address
// TESTNET NOTE: This endpoint returns empty results on Shasta Testnet for TRX transfers
// For now, we confirm payment by balance check instead of transaction lookup (see CheckPendingPayments)
// TODO: For mainnet, implement proper transaction verification with retries instead of balance checking
func (s *TronService) GetAddressTransactions(address string) ([]TronTransaction, error) {
	// Use GET request to TRC20 transactions endpoint (only returns token transfers, not native TRX)
	// NOTE: This endpoint may be empty on testnet - currently not used for balance confirmation
	url := TRON_RPC_URL + "/v1/accounts/" + address + "/transactions/trc20?limit=100"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	
	if s.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", s.apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// TESTNET DEBUG: Print transactions response (remove for production)
	fmt.Printf("[DEBUG] Transactions response: %s\n", string(body))

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var transactions []TronTransaction
	if txList, ok := result["data"].([]interface{}); ok {
		for _, tx := range txList {
			if txMap, ok := tx.(map[string]interface{}); ok {
				transaction := &TronTransaction{
					Timestamp: time.Now().Unix(),
				}
				if txID, ok := txMap["transaction_id"].(string); ok {
					transaction.TxID = txID
				}
				transactions = append(transactions, *transaction)
			}
		}
	}

	return transactions, nil
}

// ===== Helper methods =====

// callTronAPI makes a POST request to the Tron API with the given endpoint and payload
// Handles authentication and returns the raw response body or error
func (s *TronService) callTronAPI(endpoint string, payload interface{}) ([]byte, error) {
	url := TRON_RPC_URL + endpoint
	
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if s.apiKey != "" {
		req.Header.Set("TRON-PRO-API-KEY", s.apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tron api error: %d %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// addressToHex converts a Tron address from Base58 format to hexadecimal format
// TODO: Implement proper Base58 decoding (currently returns address as-is for testing)
// IMPORTANT: This stub implementation will break USDT balance checking on mainnet
func (s *TronService) addressToHex(address string) (string, error) {
	// Tron addresses start with 'T' in Base58 encoding
	// The /walletsolidity endpoints require hex format, not Base58
	// For testing, we use the address as-is since Shasta API also accepts Base58
	// TODO: For mainnet USDT, implement proper Base58 -> Hex conversion using external library
	return address, nil
}

// parseHexBalance converts a hexadecimal balance string to int64
// Handles optional '0x' prefix commonly used in blockchain APIs
func parseHexBalance(hexStr string) (int64, error) {
	// Remove '0x' prefix if present
	if len(hexStr) > 2 && hexStr[0:2] == "0x" {
		hexStr = hexStr[2:]
	}
	var balance int64
	_, err := fmt.Sscanf(hexStr, "%x", &balance)
	return balance, err
}

// toString converts various types to string representation
// Handles string, float64, and other types with appropriate conversion
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%v", val)
	default:
		return ""
	}
}
