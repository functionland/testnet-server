package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	seed          string
	authToken     string
	apiToken      string
	accessToken   string
	cleanedOrders []OrderRecord
	openSeaAPIKey string
)

type OrderRecord struct {
	OrderNo       string
	Email         string
	ShippingPhone string
	Amount        float64
}

type OrderVerificationResponse struct {
	Shipments []struct {
		DestinationAddress struct {
			ContactEmail string `json:"contact_email"`
			ContactPhone string `json:"contact_phone"`
		} `json:"destination_address"`
		OrderData struct {
			PlatformOrderNumber string `json:"platform_order_number"`
		} `json:"order_data"`
	} `json:"shipments"`
}

type FundAccountRequest struct {
	Seed   string   `json:"seed"`
	Amount *big.Int `json:"amount,omitempty"`
	To     string   `json:"to"`
}

type FundAccountResponse struct {
	Account string  `json:"account"`
	Amount  big.Int `json:"amount"`
}
type FundAccountErrorResponse struct {
	Message     string `json:"message"`
	Description string `json:"description"`
}

type IndiegogoResponse struct {
	Response []struct {
		Email string `json:"email"`
		Order struct {
			ID       int64 `json:"id"`
			Shipping struct {
				PhoneNumber string `json:"phone_number"`
			} `json:"shipping"`
		} `json:"order"`
	} `json:"response"`
}

type BalanceRequest struct {
	Account string `json:"account"`
}

type BalanceResponse struct {
	Balance big.Int `json:"balance"`
}

type BalanceErrorResponse struct {
	Message     string `json:"message"`
	Description string `json:"description"`
}

// EmailRequest represents the JSON payload structure for the Brevo API request
type EmailRequest struct {
	Sender      Sender    `json:"sender"`
	To          []ToEmail `json:"to"`
	Subject     string    `json:"subject"`
	HtmlContent string    `json:"htmlContent"`
}

// Sender represents the "sender" part of the payload
type Sender struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ToEmail represents each recipient in the "to" array
type ToEmail struct {
	Email string `json:"email"`
	Name  string `json:"name"` // This can be an empty string if the name is not used
}

type OpenSeaResponse struct {
	NFTs []struct {
		Contract string `json:"contract"`
	} `json:"nfts"`
}

const (
	easyshipAPIURL  = "https://api.easyship.com/2023-01/shipments?per_page=1&platform_order_number="
	fundAPIURL      = "https://api.node3.functionyard.fula.network/account/set_balance"
	balanceAPIURL   = "https://api.node3.functionyard.fula.network/account/balance"
	userDetailFile  = "userDetails.txt"
	openSeaNFTId    = "functional-elephants-club"
	contractAddress = "0xe44d2ce514fd50ffa3a296ee6ce01bb1ddb5b6d6"
	chain           = "matic" // Assuming the NFT is on Polygon
	streamrFile     = "streamr.txt"
)

var fundingAmount *big.Int

func (f FundAccountRequest) MarshalJSON() ([]byte, error) {
	type Alias FundAccountRequest // Create an alias to avoid infinite recursion
	return json.Marshal(&struct {
		Amount json.Number `json:"amount"` // Use json.Number for the amount
		*Alias
	}{
		Amount: json.Number(f.Amount.String()), // Convert big.Int to json.Number
		Alias:  (*Alias)(&f),
	})
}

func checkAccountBalance(accountID string) (string, error) {
	client := &http.Client{}
	balanceRequest := BalanceRequest{
		Account: accountID,
	}
	requestBody, err := json.Marshal(balanceRequest)
	if err != nil {
		log.Println("Error marshaling balance request:", err)
		return "0", err
	}

	resp, err := client.Post(balanceAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("Error sending balance request:", err)
		return "0", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading balance response body:", err)
		return "0", err
	}

	if resp.StatusCode != http.StatusOK {
		log.Println("Balance check response body:", string(bodyBytes))
		return "0", fmt.Errorf("balance check failed with status code: %d", resp.StatusCode)
	}

	var balanceResp BalanceResponse
	err = json.Unmarshal(bodyBytes, &balanceResp)
	if err != nil {
		log.Println("Error decoding balance response:", err)
		return "0", err
	}

	return balanceResp.Balance.String(), nil
}

func preprocessCSVLine(line string) string {
	// Replace all improperly quoted fields
	// For example, if the pattern is ="", replace it with the correct format
	// This is a simple example and might need to be adjusted to handle more complex cases correctly
	return strings.ReplaceAll(line, "=\"\"\"", "\"")
}

// Reads the CSV file and returns a slice of OrderRecords
func readCSVOrders(filePath string) ([]OrderRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var orders []OrderRecord
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Preprocess the line to fix any quoting issues
		processedLine := preprocessCSVLine(line)
		// Convert the processed line into a reader so it can be used by csv.NewReader
		reader := csv.NewReader(strings.NewReader(processedLine))
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err // Handle the error as appropriate
		}
		cleanedAmount := strings.Replace(strings.Trim(record[2], " $"), ",", "", -1)
		amount, _ := strconv.ParseFloat(cleanedAmount, 64) // Assuming the Amount is in the 11th column (index 10)
		orders = append(orders, OrderRecord{
			OrderNo:       strings.TrimSpace(record[0]),
			Email:         strings.TrimSpace(record[1]),
			ShippingPhone: strings.TrimSpace(record[3]),
			Amount:        amount,
		})
	}
	return orders, nil
}

func readTokensFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		seed = scanner.Text()
	}
	if scanner.Scan() {
		authToken = scanner.Text()
	}
	if scanner.Scan() {
		apiToken = scanner.Text()
	}
	if scanner.Scan() {
		accessToken = scanner.Text()
	}

	return scanner.Err()
}

func verifyNFTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Address string `json:"address"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify NFT ownership using Polygon API (you'll need to implement this)
	hasNFT := verifyNFTOwnership(data.Address)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"hasNFT": hasNFT})
}

func init() {
	var err error
	cleanedOrders, err = readCSVOrders("contributions-masked.csv")
	if err != nil {
		log.Fatalf("Error loading orders: %v", err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// Parse command-line flags
	flag.StringVar(&openSeaAPIKey, "opensea-api", "", "OpenSea API key")
	flag.Parse()

	// Check if the OpenSea API key is provided
	if openSeaAPIKey == "" {
		log.Fatal("OpenSea API key is required. Please provide it using the --opensea-api flag.")
	}

	log.Print("Server Started")
	fundingAmount, _ = new(big.Int).SetString("999999999999999999999999999999", 10)
	err := readTokensFromFile(".tokens")
	if err != nil {
		log.Fatalf("Error reading tokens: %v", err)
	}
	http.HandleFunc("/streamr", streamrHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/verify-nft", verifyNFTHandler)
	http.HandleFunc("/verify-nft-and-fund", verifyNFTAndFundHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/", registerHandler)
	log.Fatal(http.ListenAndServe(":9090", nil))
}

func readAPIKey(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read API key from file: %v", err)
	}
	return string(data), nil
}

func sendEmailDetails(toEmail string, orderID string, phoneNumber string, orderAmount float64) error {
	apiKey, err := readAPIKey("/home/ubuntu/testnet-server/brevo.key")
	if err != nil {
		log.Fatal(err)
	}

	apiURL := "https://api.brevo.com/v3/smtp/email"

	// Split the email address at "@" and use the first part as the name
	emailParts := strings.Split(toEmail, "@")
	namePart := emailParts[0] // The part before "@"

	// Construct the HTML content
	htmlContent := fmt.Sprintf(`
		<html><head></head><body>
		<p>Hello,</p>
		<p>Thank you for your request to join our network. Here are the details of your order in the system:</p>
		<ul>
			<li>Order ID: %s</li>
			<li>Phone Number: %s</li>
			<li>Order Amount: %.2f</li>
		</ul>
		<p>Please double check what you entered on the join request.</p>
		</body></html>
	`, orderID, phoneNumber, orderAmount)

	// Prepare the request payload
	emailRequest := EmailRequest{
		Sender: Sender{
			Name:  "Functionyard",              // Updated sender name
			Email: "functionyard@fula.network", // Updated sender email
		},
		To: []ToEmail{
			{
				Email: toEmail,
				Name:  namePart,
			},
		},
		Subject:     "Your Join Network Request", // Updated subject
		HtmlContent: htmlContent,
	}

	// Marshal the payload to JSON
	payloadBytes, err := json.Marshal(emailRequest)
	if err != nil {
		return fmt.Errorf("error marshaling payload to JSON: %v", err)
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set the necessary headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)

	// Make the request using the default HTTP client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request to email API: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API responded with non-OK status: %d, body: %s", resp.StatusCode, string(body))
	}

	log.Println("Email sent successfully")
	return nil
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "static/register.html")
	case "POST":

		email := r.FormValue("email")
		orderID := r.FormValue("orderId")
		phoneNumber := r.FormValue("phoneNumber")
		tokenAccountID := r.FormValue("tokenAccountId")
		appId := r.FormValue("appId")

		// Validate appId against the allowed list
		allowedAppIds := []string{"land.fx.fotos", "land.fx.blox", "main"}
		appIdValid := false
		for _, a := range allowedAppIds {
			if appId == a {
				appIdValid = true
				break
			}
		}

		if !appIdValid {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Invalid appId provided"})
			return
		}

		// Skip verifyOrder and isOrderFunded checks if appId is "land.fx.fotos"
		if appId == "land.fx.fotos" {
			email = fmt.Sprintf("random_%d@example.com", time.Now().Unix())
			orderID = fmt.Sprintf("order_%d", time.Now().Unix())
			phoneNumber = fmt.Sprintf("555-1234-%d", time.Now().Unix()%10000)
		} else {
			w.Header().Set("Content-Type", "application/json")
			if appId == "main" {
				// Check if the order has already funded 6 accounts
				fundedAccounts := getFundedAccountsCount(orderID)
				if fundedAccounts >= 6 {
					w.WriteHeader(http.StatusBadRequest)
					json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "This order has already funded the maximum number of accounts."})
					return
				}
			}
			orderFound, emailFound, foundOrderNo, foundShippingPhone, foundOrderAmount := verifyOrder(email, orderID, phoneNumber)
			if !orderFound {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Your order could not be found automatically or does not match what we have in our system. If your email is in the system you will shortly receive an email with registered order details. You can also contact testnet@fx.land"})
				if emailFound {
					err := sendEmailDetails(email, foundOrderNo, foundShippingPhone, foundOrderAmount)
					log.Println("Email sending result")
					log.Println(err)
				}
				return
			}

			if isOrderFunded(tokenAccountID, appId) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "The account is already registered. If you think this is a mistake please contact testnet@fx.land"})
				return
			}
		}

		success, errMsg := fundAccount(tokenAccountID)
		if !success {
			balance, err := checkAccountBalance(tokenAccountID)
			if err == nil && balance != "0" {
				log.Println("Account has a positive balance, considering funding successful")
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": errMsg})
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		saveUserDetails(orderID, tokenAccountID, appId)
		response := map[string]string{"status": "success", "message": "Account is funded successfully"}
		json.NewEncoder(w).Encode(response)
	default:
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Invalid request method"})
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func getFundedAccountsCount(orderID string) int {
	file, err := os.Open(userDetailFile)
	if err != nil {
		log.Println("Error opening file:", err)
		return 0
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ", ")
		if len(parts) >= 2 && parts[1] == orderID {
			count++
		}
	}
	return count
}

//lint:ignore U1000 will be used in future
func verifyOrderEasyShip(email, orderID, phoneNumber string) bool {
	client := &http.Client{}
	req, err := http.NewRequest("GET", easyshipAPIURL+orderID, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return false
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("authorization", authToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request to API:", err)
		return false
	}
	defer resp.Body.Close()

	var orderResponse OrderVerificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&orderResponse); err != nil {
		log.Println("Error decoding response:", err)
		return false
	}

	for _, shipment := range orderResponse.Shipments {
		if shipment.DestinationAddress.ContactEmail == email &&
			shipment.DestinationAddress.ContactPhone == phoneNumber &&
			shipment.OrderData.PlatformOrderNumber == orderID {
			return true
		}
	}
	return false
}

//lint:ignore U1000 will be used in future
func verifyOrderIgg(email, orderID, phoneNumber string) bool {
	// Prepare the Indiegogo API request
	client := &http.Client{}
	indiegogoAPIURL := "https://api.indiegogo.com/2/campaigns/28885449/contributions.json"
	req, err := http.NewRequest("GET", indiegogoAPIURL, nil)
	if err != nil {
		log.Println("Error creating request:", err)
		return false
	}

	// Add the required query parameters
	q := req.URL.Query()
	q.Add("api_token", apiToken)
	q.Add("access_token", accessToken)
	q.Add("email", email) // This will filter the results by the provided email
	req.URL.RawQuery = q.Encode()

	req.Header.Add("accept", "application/json")

	// Perform the API request
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request to API:", err)
		return false
	}
	defer resp.Body.Close()

	// Parse the JSON response
	var indiegogoResponse IndiegogoResponse
	if err := json.NewDecoder(resp.Body).Decode(&indiegogoResponse); err != nil {
		log.Println("Error decoding response:", err)
		return false
	}

	// Check if any order matches the provided details
	for _, contribution := range indiegogoResponse.Response {
		if contribution.Email == email &&
			fmt.Sprintf("%d", contribution.Order.ID) == orderID &&
			contribution.Order.Shipping.PhoneNumber == phoneNumber {
			return true
		}
	}

	return false
}

// Verifies the order by matching the user input against the parsed CSV records
func verifyOrder(email, orderID, phoneNumber string) (bool, bool, string, string, float64) {
	sanitizedOrderID := strings.TrimSpace(orderID)
	sanitizedEmail := strings.TrimSpace(email)
	sanitizedPhone := strings.TrimSpace(phoneNumber)
	if len(sanitizedPhone) < 4 {
		// Handle error or adjust logic as necessary
		return false, false, "", "", 0.0
	}
	sanitizedPhoneLast4 := sanitizedPhone[len(sanitizedPhone)-4:]

	emailFound := false
	foundOrderNo := ""
	foundShippingPhone := ""
	foundOrderAmount, _ := strconv.ParseFloat("0", 64)

	log.Printf("verifyOrder", "sanitizedOrderID", sanitizedOrderID, "sanitizedEmail", sanitizedEmail, "sanitizedPhone", sanitizedPhone)

	for _, order := range cleanedOrders {
		if strings.EqualFold(order.Email, sanitizedEmail) {
			emailFound = true // Email matches.
			foundOrderNo = order.OrderNo
			foundShippingPhone = order.ShippingPhone
			foundOrderAmount = order.Amount
			if len(order.ShippingPhone) < 4 {
				continue
			}
			orderPhoneLast4 := order.ShippingPhone[len(order.ShippingPhone)-4:]
			if strings.EqualFold(order.OrderNo, sanitizedOrderID) &&
				strings.EqualFold(orderPhoneLast4, sanitizedPhoneLast4) &&
				order.Amount > 1 {
				return true, true, foundOrderNo, foundShippingPhone, foundOrderAmount // Full match.
			}
		}
	}
	return false, emailFound, foundOrderNo, foundShippingPhone, foundOrderAmount // Full match not found, return status of email match.
}

func fundAccount(tokenAccountID string) (bool, string) {
	client := &http.Client{}
	fundRequest := FundAccountRequest{
		Seed:   seed,
		Amount: fundingAmount,
		To:     tokenAccountID,
	}
	requestBody, err := json.Marshal(fundRequest)
	if err != nil {
		log.Fatalf("Failed to marshal request: %v", err)
	}
	log.Println("Request body jsonData:", string(requestBody))

	resp, err := client.Post(fundAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("Error sending request to funding API:", err)
		return false, fmt.Sprintf("Error sending request to funding API: %s", err.Error())
	}
	defer resp.Body.Close()

	// Read the response body into a byte slice so it can be reused
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return false, fmt.Sprintf("Error reading response body: %s", err.Error())
	}

	var errorResp FundAccountErrorResponse
	if resp.StatusCode != http.StatusOK {
		log.Printf("Server responded with non-OK status: %d\n", resp.StatusCode)
		log.Println("Response body:", string(bodyBytes))
		errorErr := json.Unmarshal(bodyBytes, &errorResp)

		if errorErr == nil {
			// If there is no error, then it was an error response
			log.Printf("Error response from funding API: %+v\n", errorResp)
			return false, fmt.Sprintf("Error response from funding API: %+v", errorResp)
		} else {
			// If both decodes failed, there is an issue with the response format
			log.Printf("Error decoding funding response: %v\n", errorErr)
			return false, fmt.Sprintf("Error decoding funding response: %v", errorErr.Error())
		}
	}

	// Print the full response body for debugging
	log.Println("Full response body:", string(bodyBytes))

	// Attempt to decode the response into the success structure
	var fundResponse FundAccountResponse
	successErr := json.Unmarshal(bodyBytes, &fundResponse)

	if successErr == nil {
		// If there is no error, then it was a success response
		log.Printf("Funding successful: %+v\n", fundResponse)
		return fundResponse.Account == tokenAccountID && fundResponse.Amount.String() == fundingAmount.String(), ""
	}

	// Attempt to decode the response into the error structure

	errorErr := json.Unmarshal(bodyBytes, &errorResp)

	if errorErr == nil {
		// If there is no error, then it was an error response
		log.Printf("Error response from funding API: %+v\n", errorResp)
		return false, fmt.Sprintf("Error response from funding API: %+v", errorResp)
	} else {
		// If both decodes failed, there is an issue with the response format
		log.Printf("Error decoding funding response: %v\n", errorErr)
		return false, fmt.Sprintf("Error decoding funding response: %v", errorErr.Error())
	}
}

func saveUserDetails(orderID, tokenAccountID, appId string) {
	timestamp := time.Now().Format(time.RFC3339) // Get current date/time
	// Include appId in the record format
	record := fmt.Sprintf("%s, %s, %s, %s\n", timestamp, orderID, tokenAccountID, appId)

	file, err := os.OpenFile(userDetailFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(record); err != nil {
		log.Println("Error writing to file:", err)
	}
}

func isOrderFunded(tokenAccountID, appId string) bool {
	file, err := os.Open(userDetailFile)
	if err != nil {
		log.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Update parts length check for new format: "timestamp, orderID, tokenAccountID, appId"
		parts := strings.Split(line, ", ")
		if len(parts) >= 3 && parts[2] == tokenAccountID && parts[3] == appId {
			return true
		}
	}
	return false
}

func verifyNFTOwnership(address string) bool {
	url := fmt.Sprintf("https://api.opensea.io/api/v2/chain/%s/account/%s/nfts?collection=%s", chain, address, openSeaNFTId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return false
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("x-api-key", openSeaAPIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return false
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("API request failed with status code: %d\n", resp.StatusCode)
		fmt.Println("Response body:", string(body))
		return false
	}

	var openSeaResp OpenSeaResponse
	err = json.Unmarshal(body, &openSeaResp)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return false
	}

	// Check if the user owns any NFTs from the specified contract
	for _, nft := range openSeaResp.NFTs {
		if nft.Contract == contractAddress {
			return true
		}
	}

	return false
}

func verifyNFTAndFundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Address        string `json:"address"`
		TokenAccountID string `json:"tokenAccountId"`
		AppID          string `json:"appId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify NFT ownership
	hasNFT := verifyNFTOwnership(data.Address)
	if !hasNFT {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "NFT verification failed. You do not own the required NFT."})
		return
	}

	// Check if the order (address in this case) is already funded
	if isOrderFunded(data.Address, data.AppID) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "The order is already registered. If you think this is a mistake please contact testnet@fx.land"})
		return
	}

	// Fund the account
	success, errMsg := fundAccount(data.TokenAccountID)
	if !success {
		balance, err := checkAccountBalance(data.TokenAccountID)
		if err == nil && balance != "0" {
			log.Println("Account has a positive balance, considering funding successful")
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": errMsg})
			return
		}
	}

	// Save user details
	saveUserDetails(data.Address, data.TokenAccountID, data.AppID)

	// Send success response
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"status": "success", "message": "Account is funded successfully"}
	json.NewEncoder(w).Encode(response)
}

func accountExists(streamrAccount string) bool {
	file, err := os.Open(streamrFile)
	if err != nil {
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), streamrAccount) {
			return true
		}
	}
	return false
}

func sendStreamrEmail(email, orderID, phoneNumber, streamrAccount string) error {
	apiKey, err := readAPIKey("/home/ubuntu/testnet-server/brevo.key")
	if err != nil {
		return err
	}

	apiURL := "https://api.brevo.com/v3/smtp/email"

	htmlContent := fmt.Sprintf(`
        <html><head></head><body>
        <p>New Streamr node request:</p>
        <ul>
            <li>Email: %s</li>
            <li>Order ID: %s</li>
            <li>Phone Number: %s</li>
            <li>Streamr Account: %s</li>
        </ul>
        </body></html>
    `, email, orderID, phoneNumber, streamrAccount)

	emailRequest := EmailRequest{
		Sender: Sender{
			Name:  "Functionyard",
			Email: "functionyard@fula.network",
		},
		To: []ToEmail{
			{
				Email: "hi@fx.land",
				Name:  "FX Land",
			},
		},
		Subject:     "New Streamr node request",
		HtmlContent: htmlContent,
	}

	payloadBytes, err := json.Marshal(emailRequest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API responded with non-OK status: %d", resp.StatusCode)
	}

	return nil
}

func saveStreamrAccount(streamrAccount, orderID string) error {
	file, err := os.OpenFile(streamrFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%s,%s\n", streamrAccount, orderID))
	return err
}
func streamrHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "static/streamr.html")
		return
	}

	email := r.FormValue("email")
	orderID := r.FormValue("orderId")
	phoneNumber := r.FormValue("phoneNumber")
	streamrAccount := r.FormValue("streamrAccount")

	orderFound, emailFound, foundOrderNo, foundShippingPhone, foundOrderAmount := verifyOrder(email, orderID, phoneNumber)
	if !orderFound {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Your order could not be found automatically or does not match what we have in our system. If your email is in the system you will shortly receive an email with registered order details. You can also contact testnet@fx.land"})
		if emailFound {
			err := sendEmailDetails(email, foundOrderNo, foundShippingPhone, foundOrderAmount)
			log.Println("Email sending result")
			log.Println(err)
		}
		return
	}

	// Check if streamrAccount already exists in streamr.txt
	if accountExists(streamrAccount) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "This Streamr account is already registered."})
		return
	}

	// Send email to hi@fx.land
	err := sendStreamrEmail(email, orderID, phoneNumber, streamrAccount)
	if err != nil {
		log.Println("Error sending Streamr email:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Error processing your request. Please try again later."})
		return
	}

	// Save streamrAccount and orderID to streamr.txt
	err = saveStreamrAccount(streamrAccount, orderID)
	if err != nil {
		log.Println("Error saving Streamr account:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Error processing your request. Please try again later."})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success", "message": "Your Streamr node request has been submitted successfully."})
}
