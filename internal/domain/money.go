package domain

import (
	"fmt"
	"math"
)

// Money represents an exact monetary quantity stored in integer cents to prevent floating-point drift.
type Money struct {
	Cents    int64  `json:"cents"`
	Currency string `json:"currency"`
}

// NewMoney validates amount > 0 and converts decimal float to exact integer cents.
func NewMoney(amount float64, currency string) (Money, error) {
	if amount <= 0 {
		return Money{}, ErrInvalidAmount
	}
	if currency == "" {
		currency = "USD"
	}
	cents := int64(math.Round(amount * 100))
	return Money{
		Cents:    cents,
		Currency: currency,
	}, nil
}

// Format returns a human-readable string representation of the money value.
func (m Money) Format() string {
	sign := ""
	cents := m.Cents
	if cents < 0 {
		sign = "-"
		cents = -cents
	}
	return fmt.Sprintf("%s%s %.2f", sign, m.Currency, float64(cents)/100.0)
}

// Add sums two monetary values of the same currency.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.Currency, other.Currency)
	}
	return Money{
		Cents:    m.Cents + other.Cents,
		Currency: m.Currency,
	}, nil
}

// Sub subtracts other from m, verifying matching currency.
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("currency mismatch: %s vs %s", m.Currency, other.Currency)
	}
	return Money{
		Cents:    m.Cents - other.Cents,
		Currency: m.Currency,
	}, nil
}
