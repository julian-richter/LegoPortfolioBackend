package models

import (
	"time"
)

// Model is the base interface that all models must implement
type Model interface {
	GetID() int64
	SetID(id int64)
}

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        int64     `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (b BaseModel) GetID() int64 {
	return b.ID
}

func (b BaseModel) SetID(id int64) {
	b.ID = id
}
