package models

import (
	"time"

	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"users"`
	ID            int `bun:",pk,autoincrement"`
	Email         string
	Password      string
	Admin         bool
	CreatedAt     time.Time `bun:",nullzero"`
	UpdatedAt     time.Time `bun:",nullzero"`
}
