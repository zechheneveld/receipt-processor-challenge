package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProcessReceipt(t *testing.T) {
	receipt := Receipt{
		Retailer:     "Test Retailer",
		PurchaseDate: "2023-10-10",
		PurchaseTime: "14:30",
		Items: []Item{
			{ShortDescription: "Item 1", Price: 1.25},
			{ShortDescription: "Item 2", Price: 2.50},
		},
		Total: 3.75,
	}

	body, _ := json.Marshal(receipt)
	req, err := http.NewRequest("POST", "/receipts/process", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(processReceipt)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if _, ok := response["id"]; !ok {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestGetPoints(t *testing.T) {
	receipts["1"] = Receipt{
		ID:           "1",
		Retailer:     "Test Retailer",
		PurchaseDate: "2023-10-10",
		PurchaseTime: "14:30",
		Items: []Item{
			{ShortDescription: "Item 1", Price: 1.25},
			{ShortDescription: "Item 2", Price: 2.50},
		},
		Total:  3.75,
		Points: 100,
	}

	req, err := http.NewRequest("GET", "/receipts/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getPoints)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]int
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if points, ok := response["points"]; !ok || points != 100 {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}
