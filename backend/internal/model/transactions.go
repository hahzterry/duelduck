package model

type GetTransactionHistoryReq struct {
	PageNumber uint64 `json:"page_number" query:"page_number"`
	PageSize   uint64 `json:"page_size" query:"page_size"`
	RemoveSpam bool   `json:"remove_spam" query:"remove_spam"`
}

type GetTransactionHistoryResp struct {
	Transactions []Transaction     `json:"transactions"`
	Types        []TransactionType `json:"types"`
}
type Transaction struct {
	BlockID       int64  `json:"block_id"`
	BlockTime     int64  `json:"block_time"`
	TransID       string `json:"trans_id"`
	Address       string `json:"address"`
	TokenAddress  string `json:"token_address"`
	TokenAccount  string `json:"token_account"`
	TokenDecimals uint8  `json:"token_decimals"`
	Amount        int64  `json:"amount"`
	PreBalance    int64  `json:"pre_balance"`
	PostBalance   int64  `json:"post_balance"`
	ChangeType    string `json:"change_type"`
	Fee           int64  `json:"fee"`
}

type SolscanBalanceChangeResp struct {
	Data []Transaction `json:"data"`
}

type GetTokenAccountsReq struct {
	Address    string `json:"address" query:"address"`
	Type       string `json:"type" query:"type"`
	PageNumber uint64 `json:"page_number" query:"page_number"`
	PageSize   uint64 `json:"page_size" query:"page_size"`
	HideZero   bool   `json:"hide_zero" query:"hide_zero"`
}

type SolscanTokenAccountsResp struct {
	Data     []TokenAccount `json:"data"`
	Metadata struct {
		Tokens map[string]SolscanTokenMetadata `json:"tokens"`
	} `json:"metadata"`
}

type SolscanTokenMetadata struct {
	TokenAddress string `json:"token_address"`
	TokenName    string `json:"token_name"`
	TokenSymbol  string `json:"token_symbol"`
	TokenIcon    string `json:"token_icon"`
}

type TokenAccount struct {
	TokenAccount  string `json:"token_account"`
	TokenAddress  string `json:"token_address"`
	Amount        int64  `json:"amount"`
	TokenDecimals uint8  `json:"token_decimals"`
	Owner         string `json:"owner"`
}
