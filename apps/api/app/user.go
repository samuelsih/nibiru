package app

import (
	"time"
	"uuid"

	"github.com/guregu/null/v6"
)

type User struct {
	ID           uuid.UUID   `db:"id"         json:"id"        example:"01924a7d-9b3e-7c1a-8f4b-7c1d2e3f4a5b"`
	Email        string      `db:"email"      json:"email"     example:"admin@gmail.com"`
	PasswordHash string      `db:"password"   json:"-"`
	FirstName    string      `db:"first_name" json:"firstName" example:"Admin"`
	LastName     null.String `db:"last_name"  json:"lastName"  example:"New"`
	CreatedAt    time.Time   `db:"created_at" json:"createdAt" example:"2026-01-02T15:04:05Z"`
}
