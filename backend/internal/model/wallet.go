package model

import (
	"dd-prediction-api/pkg/apperrors"

	"github.com/uptrace/bun"
)

const (
	NewPdaInitBytes        = 818
	OldPdaInitBytes        = 814
	PdaMultiplierStepBytes = 330
)

var (
	ErrTokenAccountUninitialized  = apperrors.NotFound("token account is not initialized")
	ErrSolanaAccountUninitialized = "solana account is not initialized"
)

const (
	TransactionTypeDuelPrediction uint8 = 1
	TransactionTypeDuelRefund     uint8 = 2
	TransactionTypeDuelCommission uint8 = 3
	TransactionTypeDuelReward     uint8 = 4
)

type TransactionType struct {
	bun.BaseModel `bun:"table:transactions,alias:tx" json:"-"`

	Signature string `bun:"signature,pk,type:CHAR(88)" json:"signature"`
	TxType    uint8  `bun:"tx_type,type:SMALLINT,notnull" json:"tx_type"`
}

func NewTransaction(txType uint8, signature string) TransactionType {
	return TransactionType{Signature: signature, TxType: txType}
}

func NewTransactionsWithSameType(txType uint8, signatures ...string) []TransactionType {
	txs := make([]TransactionType, 0, len(signatures)+1)

	for _, signature := range signatures {
		txs = append(txs, NewTransaction(txType, signature))
	}

	return txs
}

type GetTransactionTypesReq struct {
	Signatures []string `json:"signatures"`
}

type WATxResp struct {
	RawTx []byte `json:"transaction"`
}

type WAInitTxResp struct {
	WATxResp
	RoomTokenPDA string `json:"room_token_pda"`
}

type WAPdaInfo struct {
	PdaAddress string `json:"pda_address"`
	BytesSize  uint64 `json:"bytes_size"`
}

type WASignatureResp struct {
	Signature string `json:"signature"`
}

type TokenInfo struct {
	Mint                 string      `json:"mint"`
	Standard             string      `json:"standard"`
	Name                 string      `json:"name"`
	Symbol               string      `json:"symbol"`
	Logo                 interface{} `json:"logo"`
	Decimals             string      `json:"decimals"`
	TotalSupply          string      `json:"totalSupply"`
	TotalSupplyFormatted string      `json:"totalSupplyFormatted"`
	FullyDilutedValue    string      `json:"fullyDilutedValue"`
	Metaplex             Metaplex    `json:"metaplex"`
}

type Metaplex struct {
	MetadataUri          string `json:"metadataUri"`
	MasterEdition        bool   `json:"masterEdition"`
	IsMutable            bool   `json:"isMutable"`
	PrimarySaleHappened  int    `json:"primarySaleHappened"`
	SellerFeeBasisPoints int    `json:"sellerFeeBasisPoints"`
	UpdateAuthority      string `json:"updateAuthority"`
}

type CommissionRewards struct {
	TXRecords               []TransactionType
	CreatorCommissionTxHash string
	CreatorCommissionReward uint64
}
