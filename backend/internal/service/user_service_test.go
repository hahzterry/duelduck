package service

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/mtype"
)

// ── mocks ─────────────────────────────────────────────────────────────────────

type mockUserRepository struct{ mock.Mock }

func (m *mockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	args := m.Called(ctx, id)
	res, _ := args.Get(0).(*model.User)
	return res, args.Error(1)
}

func (m *mockUserRepository) FindByProjectIDAndWallet(ctx context.Context, projectID uuid.UUID, walletAddress string) (*model.User, error) {
	args := m.Called(ctx, projectID, walletAddress)
	res, _ := args.Get(0).(*model.User)
	return res, args.Error(1)
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email mtype.Email) (*model.User, error) {
	args := m.Called(ctx, email)
	res, _ := args.Get(0).(*model.User)
	return res, args.Error(1)
}

func (m *mockUserRepository) BlockProfile(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserRepository) UnblockProfile(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserRepository) SetProjectID(ctx context.Context, userID, projectID uuid.UUID) error {
	return m.Called(ctx, userID, projectID).Error(0)
}

func (m *mockUserRepository) WithTx(_ bun.Tx) repository.IUserRepository {
	return m
}

type mockUserTransactionManager struct{ mock.Mock }

func (m *mockUserTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context, bun.Tx) error) error {
	return fn(ctx, bun.Tx{})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newTestUserService(userRepo repository.IUserRepository) *UserService {
	return &UserService{
		UserRepository:     userRepo,
		TransactionManager: &mockUserTransactionManager{},
	}
}

// ── FindByEmail ───────────────────────────────────────────────────────────────

func TestFindByEmail(t *testing.T) {
	email := mtype.Email("test@example.com")
	user := &model.User{ID: uuid.New(), Email: email}

	tests := []struct {
		name    string
		setup   func(r *mockUserRepository)
		wantErr bool
		wantNil bool
	}{
		{
			name: "user found",
			setup: func(r *mockUserRepository) {
				r.On("FindByEmail", mock.Anything, email).Return(user, nil)
			},
		},
		{
			name: "user not found returns nil",
			setup: func(r *mockUserRepository) {
				r.On("FindByEmail", mock.Anything, email).Return(nil, nil)
			},
			wantNil: true,
		},
		{
			name: "repo error propagates",
			setup: func(r *mockUserRepository) {
				r.On("FindByEmail", mock.Anything, email).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockUserRepository)
			tt.setup(repo)

			svc := newTestUserService(repo)
			got, err := svc.FindByEmail(context.Background(), email)

			if tt.wantErr {
				assert.Error(t, err)
			} else if tt.wantNil {
				require.NoError(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, user.ID, got.ID)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGetByID(t *testing.T) {
	userID := uuid.New()
	user := &model.User{ID: userID}

	tests := []struct {
		name    string
		setup   func(r *mockUserRepository)
		wantErr bool
	}{
		{
			name: "user found returns it",
			setup: func(r *mockUserRepository) {
				r.On("GetByID", mock.Anything, userID).Return(user, nil)
			},
		},
		{
			name: "user not found returns NotFound error",
			setup: func(r *mockUserRepository) {
				r.On("GetByID", mock.Anything, userID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "repo error propagates",
			setup: func(r *mockUserRepository) {
				r.On("GetByID", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(mockUserRepository)
			tt.setup(repo)

			svc := newTestUserService(repo)
			got, err := svc.GetByID(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, userID, got.ID)
			}
			repo.AssertExpectations(t)
		})
	}
}

// ── CreateWithEmail ───────────────────────────────────────────────────────────

func TestCreateWithEmail(t *testing.T) {
	email := mtype.Email("partner@example.com")

	t.Run("creates partner user without wallet address", func(t *testing.T) {
		repo := new(mockUserRepository)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
			return u.Role == mtype.RolePartner && u.Email == email
		})).Return(nil)

		svc := newTestUserService(repo)
		user, err := svc.CreateWithEmail(context.Background(), email)

		require.NoError(t, err)
		require.NotNil(t, user)
		assert.Empty(t, user.WalletAddress, "wallet address must not be generated at registration")
		repo.AssertExpectations(t)
	})

	t.Run("db error propagates", func(t *testing.T) {
		repo := new(mockUserRepository)
		repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

		svc := newTestUserService(repo)
		_, err := svc.CreateWithEmail(context.Background(), email)

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

// ── SignInWithEmail ───────────────────────────────────────────────────────────

func TestSignInWithEmail(t *testing.T) {
	email := mtype.Email("user@example.com")
	existing := &model.User{ID: uuid.New(), Email: email}

	t.Run("existing user — no creation", func(t *testing.T) {
		repo := new(mockUserRepository)
		repo.On("FindByEmail", mock.Anything, email).Return(existing, nil)

		svc := newTestUserService(repo)
		got, err := svc.SignInWithEmail(context.Background(), email)

		require.NoError(t, err)
		assert.Equal(t, existing.ID, got.ID)
		repo.AssertExpectations(t)
	})

	t.Run("new user is created without wallet address", func(t *testing.T) {
		repo := new(mockUserRepository)
		repo.On("FindByEmail", mock.Anything, email).Return(nil, nil)
		repo.On("Create", mock.Anything, mock.Anything).Return(nil)

		svc := newTestUserService(repo)
		got, err := svc.SignInWithEmail(context.Background(), email)

		require.NoError(t, err)
		assert.Empty(t, got.WalletAddress, "wallet address must not be generated at sign-in")
		repo.AssertExpectations(t)
	})

	t.Run("find error propagates", func(t *testing.T) {
		repo := new(mockUserRepository)
		repo.On("FindByEmail", mock.Anything, email).Return(nil, errors.New("db error"))

		svc := newTestUserService(repo)
		_, err := svc.SignInWithEmail(context.Background(), email)

		assert.Error(t, err)
		repo.AssertExpectations(t)
	})
}

// ── VerifySecret (Solana) ─────────────────────────────────────────────────────

func TestVerifySecret(t *testing.T) {
	svc := newTestUserService(new(mockUserRepository))

	t.Run("valid signature returns true", func(t *testing.T) {
		privKey := solana.NewWallet().PrivateKey
		address := privKey.PublicKey().String()
		sig, err := privKey.Sign([]byte(address))
		require.NoError(t, err)

		ok, err := svc.VerifySecret(context.Background(), model.AuthWithWallet{
			Address: address,
			Secret:  base64.StdEncoding.EncodeToString(sig[:]),
		})

		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("invalid address returns error", func(t *testing.T) {
		ok, err := svc.VerifySecret(context.Background(), model.AuthWithWallet{
			Address: "not-a-solana-address",
			Secret:  base64.StdEncoding.EncodeToString(make([]byte, 64)),
		})

		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("invalid base64 signature returns error", func(t *testing.T) {
		privKey := solana.NewWallet().PrivateKey
		ok, err := svc.VerifySecret(context.Background(), model.AuthWithWallet{
			Address: privKey.PublicKey().String(),
			Secret:  "!!!not-base64!!!",
		})

		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("wrong key returns false", func(t *testing.T) {
		// Sign with one key, verify with a different address.
		privKey := solana.NewWallet().PrivateKey
		otherAddress := solana.NewWallet().PrivateKey.PublicKey().String()
		sig, err := privKey.Sign([]byte(otherAddress))
		require.NoError(t, err)

		ok, err := svc.VerifySecret(context.Background(), model.AuthWithWallet{
			Address: otherAddress,
			Secret:  base64.StdEncoding.EncodeToString(sig[:]),
		})

		assert.Error(t, err)
		assert.False(t, ok)
	})
}

// ── VerifySecretWhiteBit (Ethereum) ───────────────────────────────────────────

func TestVerifySecretWhiteBit(t *testing.T) {
	svc := newTestUserService(new(mockUserRepository))

	makeEthSig := func(t *testing.T, privKeyHex string, message string) string {
		t.Helper()
		privKey, err := crypto.HexToECDSA(privKeyHex)
		require.NoError(t, err)
		msgHash := accounts.TextHash([]byte(message))
		sig, err := crypto.Sign(msgHash, privKey)
		require.NoError(t, err)
		sig[64] += 27
		return base64.StdEncoding.EncodeToString(sig)
	}

	const testPrivKey = "fad9c8855b740a0b7ed4c221dbad0f33a83a49cad6b3fe8d5817ac83d38b6a19"

	t.Run("valid ethereum signature returns true", func(t *testing.T) {
		privKey, _ := crypto.HexToECDSA(testPrivKey)
		address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
		sig := makeEthSig(t, testPrivKey, address)

		ok, err := svc.VerifySecretWhiteBit(context.Background(), model.AuthWithWallet{
			Address: address,
			Secret:  sig,
		})

		require.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("invalid Ethereum address returns error", func(t *testing.T) {
		ok, err := svc.VerifySecretWhiteBit(context.Background(), model.AuthWithWallet{
			Address: "not-an-address",
			Secret:  base64.StdEncoding.EncodeToString(make([]byte, 65)),
		})

		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("invalid base64 signature returns error", func(t *testing.T) {
		privKey, _ := crypto.HexToECDSA(testPrivKey)
		address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()

		ok, err := svc.VerifySecretWhiteBit(context.Background(), model.AuthWithWallet{
			Address: address,
			Secret:  "!!!bad-base64!!!",
		})

		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("wrong signature length returns error", func(t *testing.T) {
		privKey, _ := crypto.HexToECDSA(testPrivKey)
		address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()

		ok, err := svc.VerifySecretWhiteBit(context.Background(), model.AuthWithWallet{
			Address: address,
			Secret:  base64.StdEncoding.EncodeToString(make([]byte, 32)),
		})

		assert.Error(t, err)
		assert.False(t, ok)
	})

	t.Run("signature from wrong key returns error", func(t *testing.T) {
		privKey, _ := crypto.HexToECDSA(testPrivKey)
		address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()

		const otherKey = "2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6"
		sig := makeEthSig(t, otherKey, address)

		ok, err := svc.VerifySecretWhiteBit(context.Background(), model.AuthWithWallet{
			Address: address,
			Secret:  sig,
		})

		assert.Error(t, err)
		assert.False(t, ok)
	})
}

// ── shouldTrackActivity ───────────────────────────────────────────────────────

func TestShouldTrackActivity(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name string
		age  time.Duration
		want bool
	}{
		{"under 24h — too new", 12 * time.Hour, false},
		{"just over 24h — start of window", 24*time.Hour + time.Minute, true},
		{"15 days — mid-window", 15 * 24 * time.Hour, true},
		{"just under 30 days — end of window", 30*24*time.Hour - time.Minute, true},
		{"31 days — past window", 31 * 24 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &model.User{CreatedAt: now.Add(-tt.age)}
			assert.Equal(t, tt.want, shouldTrackActivity(user))
		})
	}
}

// ── DeriveForPath ─────────────────────────────────────────────────────────────

func TestDeriveForPath(t *testing.T) {
	seed := make([]byte, 64)
	for i := range seed {
		seed[i] = byte(i)
	}

	t.Run("valid Solana path returns 32-byte key", func(t *testing.T) {
		key, err := DeriveForPath(SolanaDerivationPath, seed)
		require.NoError(t, err)
		require.NotNil(t, key)
		assert.Len(t, key.Key, 32)
		assert.Len(t, key.ChainCode, 32)
	})

	t.Run("master path returns master key", func(t *testing.T) {
		key, err := DeriveForPath("m", seed)
		require.NoError(t, err)
		require.NotNil(t, key)
		assert.Len(t, key.Key, 32)
	})

	t.Run("empty path returns error", func(t *testing.T) {
		_, err := DeriveForPath("", seed)
		assert.Error(t, err)
	})

	t.Run("path not starting with m returns error", func(t *testing.T) {
		_, err := DeriveForPath("44'/501'/0'/0'", seed)
		assert.Error(t, err)
	})

	t.Run("non-hardened segment returns error", func(t *testing.T) {
		_, err := DeriveForPath("m/44", seed)
		assert.Error(t, err)
	})

	t.Run("same seed and path always produces same key", func(t *testing.T) {
		k1, err := DeriveForPath(SolanaDerivationPath, seed)
		require.NoError(t, err)
		k2, err := DeriveForPath(SolanaDerivationPath, seed)
		require.NoError(t, err)
		assert.Equal(t, k1.Key, k2.Key)
	})
}

// ── parseHardenedSegment ──────────────────────────────────────────────────────

func TestParseHardenedSegment(t *testing.T) {
	tests := []struct {
		name    string
		segment string
		want    uint32
		wantErr bool
	}{
		{"index 0", "0'", 0x80000000, false},
		{"index 44", "44'", 0x8000002c, false},
		{"index 501", "501'", 0x800001f5, false},
		{"non-hardened", "44", 0, true},
		{"invalid number", "abc'", 0, true},
		{"empty", "'", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHardenedSegment(tt.segment)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// ── generateMnemonic ──────────────────────────────────────────────────────────

func TestGenerateMnemonic(t *testing.T) {
	mnemonic, err := generateMnemonic()
	require.NoError(t, err)
	assert.NotEmpty(t, mnemonic)

	words := strings.Fields(mnemonic)
	assert.Equal(t, 12, len(words), "128-bit entropy should produce 12 words")

	// Two calls must produce different mnemonics.
	mnemonic2, err := generateMnemonic()
	require.NoError(t, err)
	assert.NotEqual(t, mnemonic, mnemonic2)
}

// ── generatePrivateKeyFromMnemonic ────────────────────────────────────────────

func TestGeneratePrivateKeyFromMnemonic(t *testing.T) {
	mnemonic, err := generateMnemonic()
	require.NoError(t, err)

	key, err := generatePrivateKeyFromMnemonic(mnemonic)
	require.NoError(t, err)
	require.NotNil(t, key)

	assert.Len(t, key, ed25519.PrivateKeySize)
	assert.NotEmpty(t, key.PublicKey().String())

	// Deterministic: same mnemonic → same key.
	key2, err := generatePrivateKeyFromMnemonic(mnemonic)
	require.NoError(t, err)
	assert.Equal(t, key.String(), key2.String())
}
