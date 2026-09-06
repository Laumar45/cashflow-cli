package domain_test

import (
	"errors"
	"testing"
	"time"

	"cashflow/internal/domain"
)

func TestNormalizeCategory(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr error
	}{
		{
			name:      "Uppercase with spaces and accents (Brief Requirement)",
			input:     "Comida Rápida",
			expected:  "comida-rapida",
			expectErr: nil,
		},
		{
			name:      "Accented word",
			input:     "café",
			expected:  "cafe",
			expectErr: nil,
		},
		{
			name:      "Trim whitespace",
			input:     "   almuerzo   ",
			expected:  "almuerzo",
			expectErr: nil,
		},
		{
			name:      "Hyphenated slug",
			input:     "transporte-publico",
			expected:  "transporte-publico",
			expectErr: nil,
		},
		{
			name:      "Invalid special characters",
			input:     "comida$rapida!",
			expected:  "",
			expectErr: domain.ErrInvalidCategory,
		},
		{
			name:      "Empty string",
			input:     "   ",
			expected:  "",
			expectErr: domain.ErrInvalidCategory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := domain.NormalizeCategory(tt.input)
			if tt.expectErr != nil {
				if !errors.Is(err, tt.expectErr) {
					t.Fatalf("expected error %v, got %v", tt.expectErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewTransaction(t *testing.T) {
	money, err := domain.NewMoney(15.50, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	now := time.Now()
	tx, err := domain.NewTransaction(domain.TypeExpense, "Almuerzo", money, "Menú del día", now)
	if err != nil {
		t.Fatalf("unexpected error on NewTransaction: %v", err)
	}

	if len(tx.ID) != 26 {
		t.Errorf("expected ULID length of 26 characters, got %d (ID: %s)", len(tx.ID), tx.ID)
	}
	if tx.Category != "almuerzo" {
		t.Errorf("expected normalized category 'almuerzo', got '%s'", tx.Category)
	}
	if tx.Type != domain.TypeExpense {
		t.Errorf("expected type 'expense', got '%s'", tx.Type)
	}
	if tx.Amount.Cents != 1550 {
		t.Errorf("expected 1550 cents, got %d", tx.Amount.Cents)
	}
	if tx.Description != "Menú del día" {
		t.Errorf("expected description 'Menú del día', got '%s'", tx.Description)
	}
}
