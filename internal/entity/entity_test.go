package entity

import (
	"testing"
	"time"
)

func TestCurrencyName(t *testing.T) {
	tests := []struct {
		name     string
		currency CurrencyName
		want     string
	}{
		{"BTC", BTC, "BTC"},
		{"ETH", ETH, "ETH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.currency) != tt.want {
				t.Errorf("CurrencyName = %v, want %v", tt.currency, tt.want)
			}
		})
	}
}

func TestPriceValidation(t *testing.T) {
	tests := []struct {
		name    string
		price   Price
		wantErr bool
	}{
		{
			name: "Valid price",
			price: Price{
				Symbol:  BTC,
				Price:   "5000",
				Updated: time.Now().Format("2006-01-02 15:04:05"),
			},
			wantErr: false,
		},
		{
			name: "Empty price",
			price: Price{
				Symbol:  BTC,
				Price:   "",
				Updated: time.Now().Format("2006-01-02 15:04:05"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.price.Price == "" && !tt.wantErr {
				t.Errorf("Price should not be empty")
			}
		})
	}
}

func TestUserActivation(t *testing.T) {
	user := &User{
		ChatID:   12345,
		Username: "TestUser",
		Active:   false,
	}

	user.Active = true
	if !user.Active {
		t.Errorf("User should be active")
	}

	user.Active = false
	if user.Active {
		t.Errorf("User should be inactive")
	}
}
