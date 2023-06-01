package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Item struct {
	ShortDescription string `json:"shortDescription"`
	Price            string `json:"price"`
}

type Receipt struct {
	ID           int64  `json:"id"`
	Retailer     string `json:"retailer"`
	PurchaseDate string `json:"purchaseDate"`
	PurchaseTime string `json:"purchaseTime"`
	Items        []Item `json:"items"`
	Total        string `json:"total"`
}

type PointsResponse struct {
	Points int `json:"points"`
}

type ReceiptProcessor struct {
	receipts map[int64]Receipt
}

func NewReceiptProcessor() *ReceiptProcessor {
	return &ReceiptProcessor{
		receipts: make(map[int64]Receipt),
	}
}

func (rp *ReceiptProcessor) ProcessReceipt(w http.ResponseWriter, r *http.Request) {
	var receipt Receipt
	err := json.NewDecoder(r.Body).Decode(&receipt)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	id := rp.generateID()
	receipt.ID = id
	rp.receipts[id] = receipt

	response := map[string]int64{"id": id}
	jsonResponse, err := json.Marshal(response)
	purchaseDate, err2 := time.Parse("2006-01-02", receipt.PurchaseDate)
	purchaseTime, err3 := time.Parse("15:04", receipt.PurchaseTime)
	t1 := time.Date(purchaseDate.Year(), purchaseDate.Month(), purchaseDate.Day(), purchaseTime.Hour(), purchaseTime.Minute(), purchaseTime.Second(), 0, time.UTC)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	} else if receipt.Items == nil || receipt.PurchaseTime == "" || receipt.PurchaseDate == "" || receipt.Retailer == "" || receipt.Total == "0.00" {
		http.Error(w, "The Receipt is not complete", http.StatusInternalServerError)
	} else if t1.After(time.Now()) {
		http.Error(w, "The Receipt time is after current time", http.StatusInternalServerError)
	} else if err2 != nil || err3 != nil {
		http.Error(w, "The Receipt time/date  format is wrong", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (rp *ReceiptProcessor) GetPoints(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	idStr := params["id"]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	receipt, ok := rp.receipts[id]
	if !ok {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	points := rp.calculatePoints(receipt)

	response := PointsResponse{Points: points}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (rp *ReceiptProcessor) GetReceipts(w http.ResponseWriter, r *http.Request) {
	var receipts []Receipt
	for id, receipt := range rp.receipts {
		receipt.ID = id
		receipts = append(receipts, receipt)
	}

	jsonResponse, err := json.Marshal(receipts)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (rp *ReceiptProcessor) generateID() int64 {
	return time.Now().UnixNano()
}
func (rp *ReceiptProcessor) calculateLetter(receipt Receipt) {

	// Define the regular expression pattern to match symbols and spaces
	pattern := "[^a-zA-Z0-9]+"

	// Create a regular expression object
	regex := regexp.MustCompile(pattern)

	// Replace symbols and spaces with an empty string
	replacedStr := regex.ReplaceAllString(receipt.Retailer, "")

	// Calculate the number of characters
	charCount := len(strings.TrimSpace(replacedStr))

	fmt.Println("Replaced String:", replacedStr)
	fmt.Println("Character Count:", charCount)
}
func (rp *ReceiptProcessor) calculatePoints(receipt Receipt) int {
	points := 0

	// Rule 1: One point for every alphanumeric character in the retailer name.
	pattern := "[^a-zA-Z0-9]+"
	// Create a regular expression object
	regex := regexp.MustCompile(pattern)
	// Replace symbols and spaces with an empty string
	replacedStr := regex.ReplaceAllString(receipt.Retailer, "")
	// Calculate the number of characters
	points += len(strings.TrimSpace(replacedStr))

	// Rule 2: 50 points if the total is a round dollar amount with no cents.
	total, err := strconv.ParseFloat(receipt.Total, 64)
	if err == nil && total == float64(int(total)) {
		points += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25.
	if total/0.25 == float64(int(total/0.25)) {
		points += 25
	}

	// Rule 4: 5 points for every two items on the receipt.
	points += len(receipt.Items) / 2 * 5

	// Rule 5: If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer. The result is the number of points earned.
	for _, item := range receipt.Items {
		Length := len(strings.TrimSpace(item.ShortDescription))
		if Length%3 == 0 {
			price, err := strconv.ParseFloat(item.Price, 64)
			if err == nil {
				points += int(math.Ceil(price * 0.2))
			}
		}
	}

	// Rule 6: 6 points if the day in the purchase date is odd.
	purchaseDate, err := time.Parse("2006-01-02", receipt.PurchaseDate)
	if err == nil && purchaseDate.Day()%2 == 1 {
		points += 6
	}

	// Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm.
	purchaseTime, err := time.Parse("15:04", receipt.PurchaseTime)
	if err == nil && ((purchaseTime.Hour() >= 14 && purchaseTime.Hour() < 16) || (purchaseTime.Hour() == 16 && purchaseTime.Minute() == 0)) {
		points += 10
	}

	return points
}

func main() {
	r := mux.NewRouter()
	receiptProcessor := NewReceiptProcessor()

	r.HandleFunc("/receipts/process", receiptProcessor.ProcessReceipt).Methods("POST")
	r.HandleFunc("/receipts/{id}/points", receiptProcessor.GetPoints).Methods("GET")
	r.HandleFunc("/receipts", receiptProcessor.GetReceipts).Methods("GET")

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
