package model

import (
	"time"
	"github.com/google/uuid"
)

type OauthClients struct {
	ID 			uuid.UUID `db:"id" json:"id"`
	ClientID 	string    `db:"client_id" json:"client_id"`
	ClientSecret string    `db:"client_secret" json:"client_secret"`
	Name        string    `db:"name" json:"name"`
	RedirectURI string    `db:"redirect_uri" json:"redirect_uri"`
	Scopes []string  `db:"scopes" json:"scopes"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}