package model

import (
	"dd-prediction-api/pkg/mtype"
	"time"

	"github.com/uptrace/bun"

	"github.com/google/uuid"
)

type Role uint8

type User struct {
	bun.BaseModel `bun:"table:users,alias:u" json:"-"`

	ID            uuid.UUID   `bun:",pk,type:uuid,default:uuid_generate_v4()" json:"id"`
	ProjectID     uuid.UUID   `bun:",nullzero" json:"project_id"`
	Email         mtype.Email `bun:",unique,nullzero" json:"email"`
	Role          mtype.Role  `bun:",notnull" json:"role"`
	WalletAddress string      `bun:"" json:"wallet_address"`
	IsActive      bool        `bun:",notnull" json:"is_active"`
	CreatedAt     time.Time   `bun:",notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt     time.Time   `bun:",notnull,default:current_timestamp" json:"updated_at"`
}

func NewUser(
	role mtype.Role,
) *User {
	now := time.Now().UTC()
	return &User{
		ID:        uuid.New(),
		Role:      role,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

type SignInWithEmail struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type AuthWithWallet struct {
	ProjectID uuid.UUID `json:"project_id"  binding:"required"`
	Address   string    `json:"address" binding:"required"`
	Secret    string    `json:"secret" binding:"required"`
}

type SendCode struct {
	Email string `json:"email"`
}
