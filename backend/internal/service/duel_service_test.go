package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"dd-prediction-api/internal/model"
)

// ── mocks ─────────────────────────────────────────────────────────────────────

type mockDuelTxRepo struct{ mock.Mock }

func (m *mockDuelTxRepo) BulkInsert(ctx context.Context, txs []model.DuelTransaction) error {
	return m.Called(ctx, txs).Error(0)
}

func (m *mockDuelTxRepo) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.DuelTransaction, error) {
	args := m.Called(ctx, userID)
	res, _ := args.Get(0).([]model.DuelTransaction)
	return res, args.Error(1)
}

func (m *mockDuelTxRepo) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]model.DuelTransaction, error) {
	args := m.Called(ctx, projectID)
	res, _ := args.Get(0).([]model.DuelTransaction)
	return res, args.Error(1)
}

// mockDuelTxRepoSync captures BulkInsert calls via a channel so async tests can
// wait for the goroutine without sleeping.
type mockDuelTxRepoSync struct {
	mock.Mock
	insertCh chan []model.DuelTransaction
}

func newSyncMock() *mockDuelTxRepoSync {
	return &mockDuelTxRepoSync{insertCh: make(chan []model.DuelTransaction, 1)}
}

func (m *mockDuelTxRepoSync) BulkInsert(ctx context.Context, txs []model.DuelTransaction) error {
	ret := m.Called(ctx, txs) // register first so AssertExpectations sees it
	m.insertCh <- txs
	return ret.Error(0)
}

func (m *mockDuelTxRepoSync) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.DuelTransaction, error) {
	args := m.Called(ctx, userID)
	res, _ := args.Get(0).([]model.DuelTransaction)
	return res, args.Error(1)
}

func (m *mockDuelTxRepoSync) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]model.DuelTransaction, error) {
	args := m.Called(ctx, projectID)
	res, _ := args.Get(0).([]model.DuelTransaction)
	return res, args.Error(1)
}

// waitInsert blocks until BulkInsert is called or the test times out.
func (m *mockDuelTxRepoSync) waitInsert(t *testing.T) []model.DuelTransaction {
	t.Helper()
	select {
	case txs := <-m.insertCh:
		return txs
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for BulkInsert to be called")
		return nil
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestDuelService(repo *mockDuelTxRepo) *DuelService {
	return &DuelService{DuelTxRepository: repo}
}

// ── GetFinancialHistory ───────────────────────────────────────────────────────

func TestGetFinancialHistory(t *testing.T) {
	userID := uuid.New()
	duelID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name    string
		setup   func(m *mockDuelTxRepo)
		wantLen int
		wantErr bool
		check   func(t *testing.T, entries []model.FinancialTxEntry)
	}{
		{
			name: "empty history returns empty slice",
			setup: func(m *mockDuelTxRepo) {
				m.On("GetByUserID", mock.Anything, userID).Return([]model.DuelTransaction{}, nil)
			},
			wantLen: 0,
		},
		{
			name: "repository error propagates",
			setup: func(m *mockDuelTxRepo) {
				m.On("GetByUserID", mock.Anything, userID).
					Return(nil, errors.New("clickhouse down"))
			},
			wantErr: true,
		},
		{
			name: "prediction tx maps fields correctly",
			setup: func(m *mockDuelTxRepo) {
				m.On("GetByUserID", mock.Anything, userID).Return([]model.DuelTransaction{
					{Signature: "abc123", TxType: model.TransactionTypeDuelPrediction, UserID: userID, DuelID: duelID, Amount: 0.5, CreatedAt: now},
				}, nil)
			},
			wantLen: 1,
			check: func(t *testing.T, entries []model.FinancialTxEntry) {
				e := entries[0]
				assert.Equal(t, model.TransactionTypeDuelPrediction, e.TxType)
				assert.Equal(t, "abc123", e.TxHash)
				assert.Equal(t, 0.5, e.Amount)
				assert.Equal(t, now, e.Date)
			},
		},
		{
			name: "reward tx has no hash",
			setup: func(m *mockDuelTxRepo) {
				m.On("GetByUserID", mock.Anything, userID).Return([]model.DuelTransaction{
					{Signature: "", TxType: model.TransactionTypeDuelReward, UserID: userID, DuelID: duelID, Amount: 0.9, CreatedAt: now},
				}, nil)
			},
			wantLen: 1,
			check: func(t *testing.T, entries []model.FinancialTxEntry) {
				e := entries[0]
				assert.Equal(t, model.TransactionTypeDuelReward, e.TxType)
				assert.Empty(t, e.TxHash)
				assert.Equal(t, 0.9, e.Amount)
			},
		},
		{
			name: "refund tx has no hash",
			setup: func(m *mockDuelTxRepo) {
				m.On("GetByUserID", mock.Anything, userID).Return([]model.DuelTransaction{
					{Signature: "", TxType: model.TransactionTypeDuelRefund, UserID: userID, DuelID: duelID, Amount: 0.5, CreatedAt: now},
				}, nil)
			},
			wantLen: 1,
			check: func(t *testing.T, entries []model.FinancialTxEntry) {
				assert.Equal(t, model.TransactionTypeDuelRefund, entries[0].TxType)
				assert.Empty(t, entries[0].TxHash)
			},
		},
		{
			name: "commission tx carries hash and owner amount",
			setup: func(m *mockDuelTxRepo) {
				m.On("GetByUserID", mock.Anything, userID).Return([]model.DuelTransaction{
					{Signature: "commission-hash", TxType: model.TransactionTypeDuelCommission, UserID: userID, DuelID: duelID, Amount: 0.05, CreatedAt: now},
				}, nil)
			},
			wantLen: 1,
			check: func(t *testing.T, entries []model.FinancialTxEntry) {
				e := entries[0]
				assert.Equal(t, model.TransactionTypeDuelCommission, e.TxType)
				assert.Equal(t, "commission-hash", e.TxHash)
				assert.Equal(t, 0.05, e.Amount)
			},
		},
		{
			name: "multiple mixed tx types preserve order",
			setup: func(m *mockDuelTxRepo) {
				m.On("GetByUserID", mock.Anything, userID).Return([]model.DuelTransaction{
					{Signature: "pred-1", TxType: model.TransactionTypeDuelPrediction, Amount: 0.5, CreatedAt: now},
					{Signature: "", TxType: model.TransactionTypeDuelReward, Amount: 0.9, CreatedAt: now.Add(time.Hour)},
					{Signature: "pred-2", TxType: model.TransactionTypeDuelPrediction, Amount: 0.5, CreatedAt: now.Add(2 * time.Hour)},
					{Signature: "", TxType: model.TransactionTypeDuelRefund, Amount: 0.5, CreatedAt: now.Add(3 * time.Hour)},
				}, nil)
			},
			wantLen: 4,
			check: func(t *testing.T, entries []model.FinancialTxEntry) {
				assert.Equal(t, model.TransactionTypeDuelPrediction, entries[0].TxType)
				assert.Equal(t, model.TransactionTypeDuelReward, entries[1].TxType)
				assert.Equal(t, model.TransactionTypeDuelPrediction, entries[2].TxType)
				assert.Equal(t, model.TransactionTypeDuelRefund, entries[3].TxType)
			},
		},
		{
			name: "zero amount is preserved",
			setup: func(m *mockDuelTxRepo) {
				m.On("GetByUserID", mock.Anything, userID).Return([]model.DuelTransaction{
					{Signature: "sig", TxType: model.TransactionTypeDuelPrediction, Amount: 0, CreatedAt: now},
				}, nil)
			},
			wantLen: 1,
			check: func(t *testing.T, entries []model.FinancialTxEntry) {
				assert.Equal(t, float64(0), entries[0].Amount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockDuelTxRepo{}
			tt.setup(m)

			entries, err := newTestDuelService(m).GetFinancialHistory(context.Background(), userID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, entries, tt.wantLen)
			if tt.check != nil {
				tt.check(t, entries)
			}
			m.AssertExpectations(t)
		})
	}
}

// ── logTxAsync ────────────────────────────────────────────────────────────────

func TestLogTxAsync(t *testing.T) {
	userID := uuid.New()
	duelID := uuid.New()

	tests := []struct {
		name        string
		txs         []model.DuelTransaction
		repoErr     error
		expectCall  bool
		checkStored func(t *testing.T, stored []model.DuelTransaction)
	}{
		{
			name:       "empty slice — BulkInsert never called",
			txs:        []model.DuelTransaction{},
			expectCall: false,
		},
		{
			name: "prediction tx stores correct fields",
			txs: []model.DuelTransaction{
				model.NewDuelTransaction("sig-abc", model.TransactionTypeDuelPrediction, userID, duelID, 0.5),
			},
			expectCall: true,
			checkStored: func(t *testing.T, stored []model.DuelTransaction) {
				require.Len(t, stored, 1)
				tx := stored[0]
				assert.Equal(t, "sig-abc", tx.Signature)
				assert.Equal(t, model.TransactionTypeDuelPrediction, tx.TxType)
				assert.Equal(t, userID, tx.UserID)
				assert.Equal(t, duelID, tx.DuelID)
				assert.Equal(t, 0.5, tx.Amount)
				assert.False(t, tx.CreatedAt.IsZero())
			},
		},
		{
			name: "reward tx stores empty hash and correct amount",
			txs: []model.DuelTransaction{
				model.NewDuelTransaction("", model.TransactionTypeDuelReward, userID, duelID, 0.9),
			},
			expectCall: true,
			checkStored: func(t *testing.T, stored []model.DuelTransaction) {
				require.Len(t, stored, 1)
				assert.Empty(t, stored[0].Signature)
				assert.Equal(t, model.TransactionTypeDuelReward, stored[0].TxType)
				assert.Equal(t, 0.9, stored[0].Amount)
			},
		},
		{
			name: "refund tx stores empty hash and correct amount",
			txs: []model.DuelTransaction{
				model.NewDuelTransaction("", model.TransactionTypeDuelRefund, userID, duelID, 0.5),
			},
			expectCall: true,
			checkStored: func(t *testing.T, stored []model.DuelTransaction) {
				require.Len(t, stored, 1)
				assert.Empty(t, stored[0].Signature)
				assert.Equal(t, model.TransactionTypeDuelRefund, stored[0].TxType)
				assert.Equal(t, 0.5, stored[0].Amount)
			},
		},
		{
			name: "multiple txs passed as one batch",
			txs: []model.DuelTransaction{
				model.NewDuelTransaction("sig1", model.TransactionTypeDuelPrediction, userID, duelID, 0.5),
				model.NewDuelTransaction("", model.TransactionTypeDuelReward, userID, duelID, 0.9),
			},
			expectCall: true,
			checkStored: func(t *testing.T, stored []model.DuelTransaction) {
				assert.Len(t, stored, 2)
				assert.Equal(t, "sig1", stored[0].Signature)
				assert.Empty(t, stored[1].Signature)
			},
		},
		{
			name:       "BulkInsert error is swallowed — fire and forget",
			txs:        []model.DuelTransaction{model.NewDuelTransaction("s", model.TransactionTypeDuelPrediction, userID, duelID, 1)},
			repoErr:    errors.New("clickhouse unavailable"),
			expectCall: true,
			checkStored: func(t *testing.T, stored []model.DuelTransaction) {
				assert.Len(t, stored, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSyncMock()

			if tt.expectCall {
				repo.On("BulkInsert", mock.Anything, tt.txs).Return(tt.repoErr)
			}

			svc := &DuelService{DuelTxRepository: repo}
			svc.logTxAsync(tt.txs)

			if !tt.expectCall {
				select {
				case <-repo.insertCh:
					t.Fatal("BulkInsert should not have been called for empty slice")
				case <-time.After(50 * time.Millisecond):
				}
				return
			}

			stored := repo.waitInsert(t)
			if tt.checkStored != nil {
				tt.checkStored(t, stored)
			}
			repo.AssertExpectations(t)
		})
	}
}

type mockCoinService struct{ mock.Mock }

func (m *mockCoinService) findSolanaTokenBySymbol(ctx context.Context, symbol string) (*model.SolanaToken, error) {
	args := m.Called(ctx, symbol)
	res, _ := args.Get(0).(*model.SolanaToken)
	return res, args.Error(1)
}

type mockTokenPriceRepo struct{ mock.Mock }

func (m *mockTokenPriceRepo) GetLastPrice(ctx context.Context, token string) (*model.TokenPrice, error) {
	args := m.Called(ctx, token)
	res, _ := args.Get(0).(*model.TokenPrice)
	return res, args.Error(1)
}

func (m *mockTokenPriceRepo) GetLastPricesBySymbol(ctx context.Context, symbols []string) (map[string]float64, error) {
	args := m.Called(ctx, symbols)
	res, _ := args.Get(0).(map[string]float64)
	return res, args.Error(1)
}

// ── validatePaymentType ───────────────────────────────────────────────────────

func ptr(v float64) *float64 { return &v }

func usdcToken() *model.SolanaToken {
	return &model.SolanaToken{
		Mint:       "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g",
		Symbol:     "USDC",
		Decimals:   6,
		USDPrice:   1.0,
		ProgramID:  "TokenProgramID",
		IsVerified: true,
	}
}

func setupUSDC(cs *mockCoinService, tp *mockTokenPriceRepo) {
	cs.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(usdcToken(), nil)
	tp.On("GetLastPrice", mock.Anything, usdcToken().Mint).Return(nil, nil)
}

func TestValidatePaymentType(t *testing.T) {
	// USDDuelMinJoinPrice = 0.999, USDDuelMaxJoinPrice = 5000.0
	// For USDC @ $1.0:  valid range is [0.999, 5000] tokens
	// For SOL  @ $20.5: valid range is [0.0487, 243.9] tokens

	tests := []struct {
		name      string
		symbol    string
		duelPrice float64
		priceType string
		minPrice  *float64
		maxPrice  *float64
		setup     func(cs *mockCoinService, tp *mockTokenPriceRepo)
		wantErr   bool
	}{
		{
			name:      "valid USDC price returns token info",
			symbol:    "USDC",
			duelPrice: 1.0,
			setup: func(cs *mockCoinService, tp *mockTokenPriceRepo) {
				cs.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(&model.SolanaToken{
					Mint:       "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g",
					Symbol:     "USDC",
					Decimals:   6,
					USDPrice:   1.0,
					ProgramID:  "TokenProgramID",
					IsVerified: true,
				}, nil)
				tp.On("GetLastPrice", mock.Anything, "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g").Return(nil, nil)
			},
		},
		{
			name:      "valid SOL price returns token info",
			symbol:    "SOL",
			duelPrice: 1.0,
			setup: func(cs *mockCoinService, tp *mockTokenPriceRepo) {
				cs.On("findSolanaTokenBySymbol", mock.Anything, "SOL").Return(&model.SolanaToken{
					Mint:       "So11111111111111111111111111111111111111112",
					Symbol:     "SOL",
					Decimals:   9,
					USDPrice:   20.5,
					ProgramID:  "TokenProgramID",
					IsVerified: true,
				}, nil)
				tp.On("GetLastPrice", mock.Anything, "So11111111111111111111111111111111111111112").Return(nil, nil)
			},
		},
		{
			name:      "cached price overrides token price",
			symbol:    "USDC",
			duelPrice: 1.0,
			setup: func(cs *mockCoinService, tp *mockTokenPriceRepo) {
				cs.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(&model.SolanaToken{
					Mint:      "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g",
					Symbol:    "USDC",
					Decimals:  6,
					USDPrice:  0.0, // will be overridden by cache
					ProgramID: "TokenProgramID",
				}, nil)
				tp.On("GetLastPrice", mock.Anything, "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g").Return(&model.TokenPrice{UsdPrice: 1.0}, nil)
			},
		},
		{
			name:      "duel price below minimum returns error",
			symbol:    "USDC",
			duelPrice: 0.0,
			setup: func(cs *mockCoinService, tp *mockTokenPriceRepo) {
				cs.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(&model.SolanaToken{
					Mint: "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g", Symbol: "USDC", USDPrice: 1.0,
				}, nil)
				tp.On("GetLastPrice", mock.Anything, "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g").Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:      "duel price above maximum returns error",
			symbol:    "USDC",
			duelPrice: 6000.0,
			setup: func(cs *mockCoinService, tp *mockTokenPriceRepo) {
				cs.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(&model.SolanaToken{
					Mint: "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g", Symbol: "USDC", USDPrice: 1.0,
				}, nil)
				tp.On("GetLastPrice", mock.Anything, "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g").Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:      "coin service error propagates",
			symbol:    "UNKNOWN",
			duelPrice: 1.0,
			setup: func(cs *mockCoinService, tp *mockTokenPriceRepo) {
				cs.On("findSolanaTokenBySymbol", mock.Anything, "UNKNOWN").Return(nil, errors.New("token not found"))
			},
			wantErr: true,
		},
		{
			name:      "nil token returns error",
			symbol:    "FAKE",
			duelPrice: 1.0,
			setup: func(cs *mockCoinService, tp *mockTokenPriceRepo) {
				cs.On("findSolanaTokenBySymbol", mock.Anything, "FAKE").Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name:      "token price repo error propagates",
			symbol:    "USDC",
			duelPrice: 1.0,
			setup: func(cs *mockCoinService, tp *mockTokenPriceRepo) {
				cs.On("findSolanaTokenBySymbol", mock.Anything, "USDC").Return(&model.SolanaToken{
					Mint: "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g", Symbol: "USDC", USDPrice: 1.0,
				}, nil)
				tp.On("GetLastPrice", mock.Anything, "EPjFWaLb3hqSoNE3cFJvqYvYVMnWkKZj8yFNi6BPEr5g").Return(nil, errors.New("clickhouse down"))
			},
			wantErr: true,
		},
		// ── range type ──────────────────────────────────────────────────────────
		{
			name:      "range: valid min and max",
			symbol:    "USDC",
			priceType: model.DuelPriceTypeRange,
			minPrice:  ptr(1.0),
			maxPrice:  ptr(100.0),
			setup:     setupUSDC,
		},
		{
			name:      "range: nil min_price returns error",
			symbol:    "USDC",
			priceType: model.DuelPriceTypeRange,
			minPrice:  nil,
			maxPrice:  ptr(100.0),
			setup:     setupUSDC,
			wantErr:   true,
		},
		{
			name:      "range: nil max_price returns error",
			symbol:    "USDC",
			priceType: model.DuelPriceTypeRange,
			minPrice:  ptr(1.0),
			maxPrice:  nil,
			setup:     setupUSDC,
			wantErr:   true,
		},
		{
			name:      "range: min equals max returns error",
			symbol:    "USDC",
			priceType: model.DuelPriceTypeRange,
			minPrice:  ptr(5.0),
			maxPrice:  ptr(5.0),
			setup:     setupUSDC,
			wantErr:   true,
		},
		{
			name:      "range: min greater than max returns error",
			symbol:    "USDC",
			priceType: model.DuelPriceTypeRange,
			minPrice:  ptr(50.0),
			maxPrice:  ptr(10.0),
			setup:     setupUSDC,
			wantErr:   true,
		},
		{
			name:      "range: min_price below global min returns error",
			symbol:    "USDC",
			priceType: model.DuelPriceTypeRange,
			minPrice:  ptr(0.5), // below USDDuelMinJoinPrice/1.0 = 0.999
			maxPrice:  ptr(100.0),
			setup:     setupUSDC,
			wantErr:   true,
		},
		{
			name:      "range: max_price above global max returns error",
			symbol:    "USDC",
			priceType: model.DuelPriceTypeRange,
			minPrice:  ptr(1.0),
			maxPrice:  ptr(6000.0), // above USDDuelMaxJoinPrice/1.0 = 5000
			setup:     setupUSDC,
			wantErr:   true,
		},
		{
			name:      "empty price_type defaults to fixed and validates duel_price",
			symbol:    "USDC",
			priceType: "",
			duelPrice: 1.0,
			setup:     setupUSDC,
		},
		{
			name:      "unknown price_type returns error",
			symbol:    "USDC",
			priceType: "auction",
			duelPrice: 1.0,
			setup:     setupUSDC,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coinSvc := new(mockCoinService)
			tokenPriceRepo := new(mockTokenPriceRepo)
			tt.setup(coinSvc, tokenPriceRepo)

			svc := &DuelService{
				CoinService:          coinSvc,
				TokenPriceRepository: tokenPriceRepo,
			}
			req := &model.CreateDuelReq{
				Symbol:    tt.symbol,
				DuelPrice: tt.duelPrice,
				PriceType: tt.priceType,
				MinPrice:  tt.minPrice,
				MaxPrice:  tt.maxPrice,
			}

			tokenInfo, err := svc.validatePaymentType(req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, tokenInfo)
			} else {
				require.NoError(t, err)
				require.NotNil(t, tokenInfo)
				assert.Equal(t, tt.symbol, tokenInfo.Symbol)
			}
			coinSvc.AssertExpectations(t)
			tokenPriceRepo.AssertExpectations(t)
		})
	}
}

// ── price range reward math ───────────────────────────────────────────────────

// TestPriceRangeRewardMath tests the proportional reward calculation formula
// used in resolveCryptoDuel for price_range duels, isolated from DB/chain deps.
//
// formula:
//
//	commissionRaw = totalPool * commissionRate/100 * priceMultiplier
//	netPool       = totalPool * (1 - commissionRate/100)
//	winner.Amount = netPool * (winner.PaidPrice / winnersPool) * priceMultiplier
func TestPriceRangeRewardMath(t *testing.T) {
	const multiplier = 1_000_000.0 // USDC 6 decimals

	calcAmounts := func(totalPool float64, commissionRate uint64, winners []model.CryptoDuelPlayer) (commissionRaw uint64, amounts []uint64) {
		commissionFraction := float64(commissionRate) / 100.0
		netPool := totalPool * (1.0 - commissionFraction)
		commissionRaw = uint64(totalPool * commissionFraction * multiplier)

		var winnersPool float64
		for _, w := range winners {
			if w.PaidPrice != nil {
				winnersPool += *w.PaidPrice
			}
		}

		amounts = make([]uint64, len(winners))
		for i, w := range winners {
			paid := 0.0
			if w.PaidPrice != nil {
				paid = *w.PaidPrice
			}
			amounts[i] = uint64(netPool * (paid / winnersPool) * multiplier)
		}
		return
	}

	tests := []struct {
		name           string
		totalPool      float64
		commissionRate uint64
		winners        []model.CryptoDuelPlayer
		wantCommission uint64
		wantAmounts    []uint64
	}{
		{
			name:           "equal stakes split net pool evenly",
			totalPool:      10.0, // 2 winners @ 2.0, 2 losers @ 3.0
			commissionRate: 10,
			winners: []model.CryptoDuelPlayer{
				{PaidPrice: ptr(2.0)},
				{PaidPrice: ptr(2.0)},
			},
			wantCommission: 1_000_000,                      // 10% of 10 USDC
			wantAmounts:    []uint64{4_500_000, 4_500_000}, // 9/2 each
		},
		{
			name:           "proportional: 1:3 stake ratio",
			totalPool:      10.0, // 2 winners @ 1.0 and 3.0, 2 losers @ 3.0
			commissionRate: 0,
			winners: []model.CryptoDuelPlayer{
				{PaidPrice: ptr(1.0)}, // 1/(1+3) = 25% → 2.5 USDC
				{PaidPrice: ptr(3.0)}, // 3/(1+3) = 75% → 7.5 USDC
			},
			wantCommission: 0,
			wantAmounts:    []uint64{2_500_000, 7_500_000},
		},
		{
			name:           "single winner takes entire net pool",
			totalPool:      10.0,
			commissionRate: 5,
			winners: []model.CryptoDuelPlayer{
				{PaidPrice: ptr(2.0)},
			},
			wantCommission: 500_000,             // 5% of 10
			wantAmounts:    []uint64{9_500_000}, // all of net pool
		},
		{
			name:           "zero commission passes full pool to winners",
			totalPool:      6.0,
			commissionRate: 0,
			winners: []model.CryptoDuelPlayer{
				{PaidPrice: ptr(1.0)}, // 1/3
				{PaidPrice: ptr(2.0)}, // 2/3
			},
			wantCommission: 0,
			wantAmounts:    []uint64{2_000_000, 4_000_000},
		},
		{
			name:           "max commission rate",
			totalPool:      10.0,
			commissionRate: 20, // MaxDuelCommissionRate
			winners: []model.CryptoDuelPlayer{
				{PaidPrice: ptr(2.0)},
				{PaidPrice: ptr(2.0)},
			},
			wantCommission: 2_000_000,                      // 20% of 10
			wantAmounts:    []uint64{4_000_000, 4_000_000}, // 8/2 each
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commissionRaw, amounts := calcAmounts(tt.totalPool, tt.commissionRate, tt.winners)

			assert.Equal(t, tt.wantCommission, commissionRaw, "commission mismatch")
			require.Len(t, amounts, len(tt.wantAmounts))
			for i, want := range tt.wantAmounts {
				assert.Equal(t, want, amounts[i], "winner[%d] amount mismatch", i)
			}

			// invariant: sum of rewards must not exceed net pool (rounding losses ok)
			commissionFraction := float64(tt.commissionRate) / 100.0
			netPoolRaw := uint64(tt.totalPool * (1.0 - commissionFraction) * multiplier)
			var totalPaid uint64
			for _, a := range amounts {
				totalPaid += a
			}
			assert.LessOrEqual(t, totalPaid, netPoolRaw, "winners received more than net pool")
		})
	}
}
