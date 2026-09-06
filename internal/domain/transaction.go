package domain

import (
	"crypto/rand"
	"regexp"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type TransactionType string

const (
	TypeIncome  TransactionType = "income"
	TypeExpense TransactionType = "expense"
)

var categoryRegex = regexp.MustCompile(`^[a-z0-9-_]+$`)

// NormalizeCategory normalizes category names: trims whitespace, replaces spaces with hyphens,
// removes diacritics, and validates against [a-z0-9-_]+.
func NormalizeCategory(raw string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(raw))
	
	// Replace accented characters commonly found in domain usage
	replacer := strings.NewReplacer(
		"á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u",
		"à", "a", "è", "e", "ì", "i", "ò", "o", "ù", "u",
		"ä", "a", "ë", "e", "ï", "i", "ö", "o", "ü", "u",
		"ñ", "n",
	)
	slug = replacer.Replace(slug)
	slug = strings.ReplaceAll(slug, " ", "-")

	if slug == "" || !categoryRegex.MatchString(slug) {
		return "", ErrInvalidCategory
	}
	return slug, nil
}

// Transaction represents an immutable financial event record.
type Transaction struct {
	ID          string          `json:"id"`
	Timestamp   time.Time       `json:"timestamp"`
	Type        TransactionType `json:"type"`
	Category    string          `json:"category"`
	Amount      Money           `json:"amount"`
	Description string          `json:"description,omitempty"`
}

// NewTransaction constructs and validates a new immutable Transaction entity.
func NewTransaction(txType TransactionType, rawCategory string, amount Money, desc string, ts time.Time) (Transaction, error) {
	if txType != TypeIncome && txType != TypeExpense {
		return Transaction{}, ErrInvalidAmount
	}
	if amount.Cents <= 0 {
		return Transaction{}, ErrInvalidAmount
	}
	category, err := NormalizeCategory(rawCategory)
	if err != nil {
		return Transaction{}, err
	}

	if ts.IsZero() {
		ts = time.Now().UTC()
	} else {
		ts = ts.UTC()
	}

	entropy := ulid.Monotonic(rand.Reader, 0)
	id, err := ulid.New(ulid.Timestamp(ts), entropy)
	if err != nil {
		return Transaction{}, err
	}

	return Transaction{
		ID:          id.String(),
		Timestamp:   ts,
		Type:        txType,
		Category:    category,
		Amount:      amount,
		Description: strings.TrimSpace(desc),
	}, nil
}
