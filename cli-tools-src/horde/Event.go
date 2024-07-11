package horde

import (
	"github.com/google/uuid"
	"time"
)

type Event struct {
	Id          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	ValidSince  time.Time `json:"validSince"`
	ValidUntil  time.Time `json:"validUntil"`
	Description *string   `json:"description,omitempty"`
	LimitedTo   []string  `json:"limitedTo,omitempty"`
	Link        *string   `json:"link,omitempty"`
	Channels    []string  `json:"channels,omitempty"`
}
