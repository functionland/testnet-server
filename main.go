package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	seed      string
	authToken string
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
	Seed   string `json:"seed"`
	Amount string `json:"amount"`
	To     string `json:"to"`
}

type FundAccountResponse struct {
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

const (
	easyshipAPIURL = "https://api.easyship.com/2023-01/shipments?per_page=1&platform_order_number="
	fundAPIURL     = "https://127.0.0.1:4000/account/fund"
	userDetailFile = "userDetails.txt"
	fundingAmount  = "1000000000000000000"
)

// Reads the CSV file and returns a slice of OrderRecords
func readCSVOrders(filePath string) ([]OrderRecord, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t' // Set the delimiter to tab if your data is tab-delimited
	reader.TrimLeadingSpace = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var orders []OrderRecord
	for _, record := range records[1:] { // Skip header row
		amount, _ := strconv.ParseFloat(strings.Trim(record[10], " $"), 64) // Assuming the Amount is in the 11th column (index 10)
		orders = append(orders, OrderRecord{
			OrderNo:       strings.TrimSpace(record[0]),
			Email:         strings.TrimSpace(record[8]),
			ShippingPhone: strings.TrimSpace(record[16]),
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

	return scanner.Err()
}

func main() {
	fmt.Print("Server Started")
	err := readTokensFromFile(".tokens")
	if err != nil {
		log.Fatalf("Error reading tokens: %v", err)
	}
	http.HandleFunc("/register", registerHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Fatal(http.ListenAndServe(":9090", nil))
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

		w.Header().Set("Content-Type", "application/json")

		if !verifyOrder(email, orderID, phoneNumber) {
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Your order could not be found automatically, please contact testnet@fx.land"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if isAccountFunded(tokenAccountID) {
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "The account is already funded. If you think this is a mistake please contact testnet@fx.land"})
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !fundAccount(tokenAccountID) {
			json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Account details were found but there was an issue funding the account. Please try again in a few minutes or contact the support at testnet@fx.land"})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		saveUserDetails(tokenAccountID)
		response := map[string]string{"status": "success", "message": "Account is funded successfully"}
		json.NewEncoder(w).Encode(response)
	default:
		json.NewEncoder(w).Encode(map[string]string{"status": "error", "message": "Invalid request method"})
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
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

// Verifies the order by matching the user input against the parsed CSV records
func verifyOrder(orderID, email, phoneNumber string) bool {
	orders, err := readCSVOrders("contributions.csv")
	if err != nil {
		log.Println("Error reading CSV file:", err)
		return false
	}

	// Sanitize the user input
	sanitizedOrderID := strings.TrimSpace(orderID)
	sanitizedEmail := strings.TrimSpace(email)
	sanitizedPhone := strings.TrimSpace(phoneNumber)

	// Search for a matching record
	for _, order := range orders {
		if order.OrderNo == sanitizedOrderID &&
			order.Email == sanitizedEmail &&
			order.ShippingPhone == sanitizedPhone &&
			order.Amount > 1 {
			return true
		}
	}
	return false
}

func fundAccount(tokenAccountID string) bool {
	client := &http.Client{}
	fundRequest := FundAccountRequest{
		Seed:   seed,
		Amount: fundingAmount,
		To:     tokenAccountID,
	}
	requestBody, err := json.Marshal(fundRequest)
	if err != nil {
		log.Println("Error marshaling request:", err)
		return false
	}

	resp, err := client.Post(fundAPIURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		log.Println("Error sending request to funding API:", err)
		return false
	}
	defer resp.Body.Close()

	var fundResponse FundAccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&fundResponse); err != nil {
		log.Println("Error decoding funding response:", err)
		return false
	}

	return fundResponse.To == tokenAccountID && fmt.Sprintf("%d", fundResponse.Amount) == fundingAmount
}

func saveUserDetails(tokenAccountID string) {
	file, err := os.OpenFile(userDetailFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	if _, err := file.WriteString(tokenAccountID + "\n"); err != nil {
		log.Println("Error writing to file:", err)
	}
}

func isAccountFunded(tokenAccountID string) bool {
	file, err := os.Open(userDetailFile)
	if err != nil {
		log.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() == tokenAccountID {
			return true
		}
	}
	return false
}
