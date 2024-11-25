package main

import (
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Receipt struct {
	ID           string  `json:"id"`
	Retailer     string  `json:"retailer"`
	PurchaseDate string  `json:"purchaseDate"`
	PurchaseTime string  `json:"purchaseTime"`
	Items        []Item  `json:"items"`
	Total        float64 `json:"total,string"`
	Points       int     `json:"points"`
}

type Item struct {
	ShortDescription string  `json:"shortDescription"`
	Price            float64 `json:"price,string"`
}

var receipts = make(map[string]Receipt)

func main() {
	http.HandleFunc("/receipts/process", processReceipt)
	http.HandleFunc("/receipts/", getPoints)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}

func processReceipt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var receipt Receipt
	if err := json.NewDecoder(r.Body).Decode(&receipt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	receipt.ID = uuid.New().String()
	receipt.Points = calculatePoints(receipt)
	receipts[receipt.ID] = receipt

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": receipt.ID})
}

func getPoints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/receipts/")
	id = strings.TrimSuffix(id, "/points")
	receipt, exists := receipts[id]
	if !exists {
		http.Error(w, "Receipt not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"points": receipt.Points})
}

func calculatePoints(receipt Receipt) int {
	points := 0

	// Rule 1: One point for every alphanumeric character in the retailer name
	alphanumeric := regexp.MustCompile(`[a-zA-Z0-9]`)
	rule1Points := len(alphanumeric.FindAllString(receipt.Retailer, -1))
	points += rule1Points
	log.Printf("Rule 1: %d points", rule1Points)

	// Rule 2: 50 points if the total is a round dollar amount with no cents
	total := receipt.Total
	rule2Points := 0
	if total == float64(int(total)) {
		rule2Points = 50
		points += rule2Points
	}
	log.Printf("Rule 2: %d points", rule2Points)

	// Rule 3: 25 points if the total is a multiple of 0.25
	rule3Points := 0
	if math.Mod(total, 0.25) == 0 {
		rule3Points = 25
		points += rule3Points
	}
	log.Printf("Rule 3: %d points", rule3Points)

	// Rule 4: 5 points for every two items on the receipt
	rule4Points := (len(receipt.Items) / 2) * 5
	points += rule4Points
	log.Printf("Rule 4: %d points", rule4Points)

	// Rule 5: Points for item descriptions that are a multiple of 3
	rule5Points := 0
	for _, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc)%3 == 0 {
			price := item.Price
			itemPoints := int(math.Ceil(price * 0.2))
			rule5Points += itemPoints
		}
	}
	points += rule5Points
	log.Printf("Rule 5: %d points", rule5Points)

	// Rule 6: 6 points if the day in the purchase date is odd
	purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
	rule6Points := 0
	if purchaseDate.Day()%2 != 0 {
		rule6Points = 6
		points += rule6Points
	}
	log.Printf("Rule 6: %d points", rule6Points)

	// Rule 7: 10 points if the time of purchase is after 2:00pm and before 4:00pm
	purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime)
	rule7Points := 0
	if purchaseTime.Hour() == 14 {
		rule7Points = 10
		points += rule7Points
	}
	log.Printf("Rule 7: %d points", rule7Points)

	log.Printf("Total points: %d", points)
	return points
}
