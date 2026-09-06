package domain_test

import (
	"errors"
	"testing"

	"cashflow/internal/domain"
)

func TestNewMoney(t *testing.T) {
	tests := []struct {
		name        string
		amount      float64
		currency    string
		expectCents int64
		expectErr   error
	}{
		{
			name:        "Standard positive amount",
			amount:      12.50,
			currency:    "USD",
			expectCents: 1250,
			expectErr:   nil,
		},
		{
			name:        "Default currency when empty",
			amount:      3000.00,
			currency:    "",
			expectCents: 300000,
			expectErr:   nil,
		},
		{
			name:        "Cent precision rounding",
			amount:      0.01,
			currency:    "USD",
			expectCents: 1,
			expectErr:   nil,
		},
		{
			name:        "Zero amount is rejected",
			amount:      0.0,
			currency:    "USD",
			expectCents: 0,
			expectErr:   domain.ErrInvalidAmount,
		},
		{
			name:        "Negative amount is rejected",
			amount:      -15.50,
			currency:    "USD",
			expectCents: 0,
			expectErr:   domain.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := domain.NewMoney(tt.amount, tt.currency)
			if tt.expectErr != nil {
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if m.Cents != tt.expectCents {
				t.Errorf("expected cents %d, got %d", tt.expectCents, m.Cents)
			}
		})
	}
}

func TestMoneyFormat(t *testing.T) {
	m, err := domain.NewMoney(25.50, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "USD 25.50"
	if formatted := m.Format(); formatted != expected {
		t.Errorf("expected format '%s', got '%s'", expected, formatted)
	}

	neg := domain.Money{Cents: -500, Currency: "USD"}
	expectedNeg := "-USD 5.00"
	if formatted := neg.Format(); formatted != expectedNeg {
		t.Errorf("expected format '%s', got '%s'", expectedNeg, formatted)
	}
}

func TestMoneyAddAndSub(t *testing.T) {
	m1, _ := domain.NewMoney(10.00, "USD")
	m2, _ := domain.NewMoney(5.50, "USD")

	sum, err := m1.Add(m2)
	if err != nil {
		t.Fatalf("unexpected error on Add: %v", err)
	}
	if sum.Cents != 1550 {
		t.Errorf("expected sum 1550 cents, got %d", sum.Cents)
	}

	diff, err := m1.Sub(m2)
	if err != nil {
		t.Fatalf("unexpected error on Sub: %v", err)
	}
	if diff.Cents != 450 {
		t.Errorf("expected diff 450 cents, got %d", diff.Cents)
	}

	eur, _ := domain.NewMoney(10.00, "EUR")
	if _, err := m1.Add(eur); err == nil {
		t.Errorf("expected error when adding different currencies, got nil")
	}
}
