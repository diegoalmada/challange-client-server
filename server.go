package main

import (
	"context"
	"encoding/json"
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const ExchangeRateUrl = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

type CurrencyData struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	Bid        string `json:"bid"`
	CreateDate string `json:"create_date"`
}

type ExchangeRate struct {
	Data CurrencyData `json:"USDBRL"`
}

type Quotation struct {
	ID         int       `gorm:"primaryKey" json:"-"`
	Code       string    `json:"code"`
	Name       string    `json:"name"`
	Bid        float64   `json:"bid"`
	QuotatedAt time.Time `json:"quotated_at"`
	gorm.Model `json:"-"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", handler)
	http.ListenAndServe(":8080", mux)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	rate, err := findExchangeRate(ctx)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Println("Exchange rate request has been canceled.")
			status = http.StatusRequestTimeout
		}
		http.Error(w, http.StatusText(status), status)
		return
	}

	quotation, err := newQuotation(rate)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = saveQuotation(ctx, quotation)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			log.Println("DB Store has been canceled.")
			status = http.StatusRequestTimeout
		}
		http.Error(w, http.StatusText(status), http.StatusRequestTimeout)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(quotation)
}

func findExchangeRate(ctx context.Context) (*CurrencyData, error) {
	reqCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, ExchangeRateUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rate ExchangeRate
	err = json.Unmarshal(body, &rate)
	if err != nil {
		return nil, err
	}
	return &rate.Data, nil
}

func saveQuotation(ctx context.Context, quotation *Quotation) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	db, err := gorm.Open(sqlite.Open("app.db"), &gorm.Config{})
	if err != nil {
		return err
	}
	db.WithContext(ctx).AutoMigrate(&Quotation{})

	result := db.WithContext(ctx).Create(quotation)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func newQuotation(currency *CurrencyData) (*Quotation, error) {
	bid, err := strconv.ParseFloat(currency.Bid, 64)
	if err != nil {
		return nil, err
	}

	layout := "2006-01-02 15:04:05"
	date, err := time.Parse(layout, currency.CreateDate)
	if err != nil {
		return nil, err
	}

	quo := &Quotation{
		Name:       currency.Name,
		Code:       currency.Code,
		Bid:        bid,
		QuotatedAt: date,
	}

	return quo, nil
}
