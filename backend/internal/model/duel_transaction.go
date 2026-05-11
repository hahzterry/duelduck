package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/go-clickhouse/ch"
)

type DuelTransaction struct {
	ch.CHModel `ch:"table:duel_transactions"`

	Signature string    `ch:"signature"`
	TxType    uint8     `ch:"tx_type"`
	UserID    uuid.UUID `ch:"user_id"`
	DuelID    uuid.UUID `ch:"duel_id"`
	ProjectID uuid.UUID `ch:"project_id"`
	Amount    float64   `ch:"amount"`
	CreatedAt time.Time `ch:"created_at"`
}

func NewDuelTransaction(signature string, txType uint8, userID, duelID uuid.UUID, amount float64) DuelTransaction {
	return DuelTransaction{
		Signature: signature,
		TxType:    txType,
		UserID:    userID,
		DuelID:    duelID,
		CreatedAt: time.Now().UTC(),
		Amount:    amount,
	}
}

func NewDuelTransactionWithProject(signature string, txType uint8, userID, duelID, projectID uuid.UUID, amount float64) DuelTransaction {
	return DuelTransaction{
		Signature: signature,
		TxType:    txType,
		UserID:    userID,
		DuelID:    duelID,
		ProjectID: projectID,
		Amount:    amount,
		CreatedAt: time.Now().UTC(),
	}
}
