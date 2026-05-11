package service

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"dd-prediction-api/config"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/internal/storage/cypher"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apperrors"
	"dd-prediction-api/pkg/mtype"
	repo "dd-prediction-api/pkg/repository"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
	"github.com/tyler-smith/go-bip39"
	"github.com/uptrace/bun"
)

type UserService struct {
	UserRepository       repository.IUserRepository
	AuthService          *AuthService
	privateKeyRepository *cypher.PrivateKeyRepository
	TransactionManager   repo.ITransactionManager
}

func NewUserService(
	c *config.Config,
	userRepository *repository.UserRepository,
	authService *AuthService,
	privateKeyRepository *cypher.PrivateKeyRepository,
	transactionManager *repo.TransactionManager,
) *UserService {
	return &UserService{
		UserRepository:       userRepository,
		AuthService:          authService,
		privateKeyRepository: privateKeyRepository,
		TransactionManager:   transactionManager,
	}
}

func (s *UserService) CreateWithEmail(
	ctx context.Context,
	email mtype.Email,
) (*model.User, error) {

	user := model.NewUser(mtype.RolePartner)

	user.Email = email

	err := s.UserRepository.Create(ctx, user)
	if err != nil {
		return nil, apperrors.Internal("failed to create user", err)
	}

	return user, nil
}

func (s *UserService) FindByEmail(ctx context.Context, email mtype.Email) (*model.User, error) {
	user, err := s.UserRepository.FindByEmail(ctx, email)
	if err != nil {
		return nil, apperrors.Internal("failed to find user by email", err)
	}

	return user, nil
}

func (s *UserService) SignInWithEmail(
	ctx context.Context,
	email mtype.Email,
) (*model.User, error) {
	user, err := s.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = s.CreateWithEmail(ctx, email)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) SignInWithGoogle(
	ctx context.Context,
	email mtype.Email,
	googleImageURL string,
) (*model.User, error) {
	user, err := s.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = s.CreateWithEmail(
			ctx,
			email,
		)
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (s *UserService) SignInWithWallet(
	ctx context.Context,
	authWallet model.AuthWithWallet,
) (*model.User, error) {
	user, err := s.GetByPublicAddress(ctx, authWallet)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = s.CreateWithWallet(ctx, &authWallet)
		if err != nil {
			return nil, err
		}
	}

	if shouldTrackActivity(user) {
		click.GetActivityBuffer().Add(model.UserActivity{
			UserID:        user.ID,
			ActivityType:  model.ActivityLogin,
			CreatedAt:     time.Now().UTC(),
			ActivityCount: 1,
			Metadata:      loginMetadata("wallet"),
		})
	}

	return user, nil
}

func (s *UserService) GetByPublicAddress(
	ctx context.Context,
	authWallet model.AuthWithWallet,
) (*model.User, error) {
	ok, err := s.VerifySecret(ctx, authWallet)
	if err != nil || !ok {
		return nil, err
	}

	user, err := s.UserRepository.FindByProjectIDAndWallet(ctx, authWallet.ProjectID, authWallet.Address)
	if err != nil {
		return nil, apperrors.Internal("failed to find user by project id and wallet address", err)
	}

	return user, nil
}

func (s *UserService) VerifySecret(
	ctx context.Context,
	authWallet model.AuthWithWallet,
) (bool, error) {
	publicKey, err := solana.PublicKeyFromBase58(authWallet.Address)
	if err != nil {
		return false, apperrors.BadRequest("invalid solana address", err)
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(authWallet.Secret)
	if err != nil {
		return false, apperrors.BadRequest("invalid signature format", err)
	}

	var signature solana.Signature
	copy(signature[:], signatureBytes)

	verified := signature.Verify(publicKey, []byte(authWallet.Address))
	if !verified {
		return false, apperrors.BadRequest("invalid signature")
	}

	return true, nil
}

func (s *UserService) VerifySecretWhiteBit(
	ctx context.Context,
	authWallet model.AuthWithWallet,
) (bool, error) {
	if !common.IsHexAddress(authWallet.Address) {
		return false, apperrors.BadRequest("invalid Ethereum address")
	}

	address := common.HexToAddress(authWallet.Address)

	signature, err := base64.StdEncoding.DecodeString(authWallet.Secret)
	if err != nil {
		return false, apperrors.BadRequest("invalid signature base64 format", err)
	}

	if len(signature) != 65 {
		return false, apperrors.BadRequest("invalid signature length")
	}

	if signature[64] >= 27 {
		signature[64] -= 27
	}

	msgHash := accounts.TextHash([]byte(authWallet.Address))

	pubKey, err := crypto.SigToPub(msgHash, signature)
	if err != nil {
		return false, apperrors.BadRequest("signature verification failed", err)
	}

	if crypto.PubkeyToAddress(*pubKey) != address {
		return false, apperrors.BadRequest("invalid signature", nil)
	}

	return true, nil
}

func (s *UserService) CreateWithWallet(
	ctx context.Context,
	authWallet *model.AuthWithWallet,
) (*model.User, error) {

	user := model.NewUser(mtype.RoleUser)
	user.WalletAddress = authWallet.Address
	user.ProjectID = authWallet.ProjectID

	err := s.TransactionManager.WithinTransaction(ctx, func(ctx context.Context, tx bun.Tx) error {
		err := s.UserRepository.WithTx(tx).Create(ctx, user)
		if err != nil {
			return apperrors.Internal("failed to create user", err)
		}

		user, err = s.UserRepository.WithTx(tx).GetByID(ctx, user.ID)
		if err != nil {
			return apperrors.Internal("failed to reload user after creation", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.User, error) {
	user, err := s.UserRepository.GetByID(ctx, id)
	if err != nil {
		return nil, apperrors.Internal("failed to get user by id", err)
	}
	if user == nil {
		return nil, apperrors.NotFound("user not found")
	}

	return user, nil
}

func (s *UserService) generateCustodialAddress(ctx context.Context, userID uuid.UUID) (string, error) {
	mnemonic, err := generateMnemonic()
	if err != nil {
		return "", err
	}

	privateKey, err := generatePrivateKeyFromMnemonic(mnemonic)
	if err != nil || privateKey == nil {
		return "", apperrors.Internal("failed to generate wallet from mnemonic", err)
	}

	if err = s.privateKeyRepository.WritePrivateKey(ctx, userID, mnemonic, ed25519.PrivateKey(privateKey)); err != nil {
		return "", apperrors.Internal("failed to store private key", err)
	}

	return privateKey.PublicKey().String(), nil
}
func generateMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", apperrors.Internal("failed to generate an entropy", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", apperrors.Internal("failed to generate a mnemonic phrase", err)
	}

	return mnemonic, nil
}

func generatePrivateKeyFromMnemonic(mnemonic string) (solana.PrivateKey, error) {
	seed := bip39.NewSeed(mnemonic, "")

	privateKey, err := deriveAccount(seed)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

const (
	SolanaDerivationPath = "m/44'/501'/0'/0'"
)

func deriveAccount(seed []byte) (solana.PrivateKey, error) {
	derivedPrivateKey, err := deriveEd25519Key(seed, SolanaDerivationPath)
	if err != nil {
		return nil, err
	}

	return solana.PrivateKey(derivedPrivateKey), nil
}

func deriveEd25519Key(seed []byte, derivationPath string) (ed25519.PrivateKey, error) {
	masterKey, err := DeriveForPath(derivationPath, seed)
	if err != nil {
		return nil, apperrors.Internal("failed to derive a seed", err)
	}

	return ed25519.NewKeyFromSeed(masterKey.Key), nil
}

type ExtendedKey struct {
	Key       []byte
	ChainCode []byte
}

func DeriveForPath(path string, seed []byte) (*ExtendedKey, error) {
	if path == "" || path[0] != 'm' {
		return nil, fmt.Errorf("invalid derivation path: %s", path)
	}

	I := hmacSHA512([]byte("ed25519 seed"), seed)
	if len(I) != 64 {
		return nil, fmt.Errorf("invalid master key length: %d", len(I))
	}

	key := I[:32]
	chainCode := I[32:]

	if path == "m" {
		return &ExtendedKey{
			Key:       append([]byte(nil), key...),
			ChainCode: append([]byte(nil), chainCode...),
		}, nil
	}

	segments := strings.Split(path, "/")[1:]
	for _, segment := range segments {
		index, err := parseHardenedSegment(segment)
		if err != nil {
			return nil, err
		}

		data := make([]byte, 0, 1+32+4)
		data = append(data, 0x00)
		data = append(data, key...)

		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], index)
		data = append(data, buf[:]...)

		I = hmacSHA512(chainCode, data)
		if len(I) != 64 {
			return nil, fmt.Errorf("invalid derived key length: %d", len(I))
		}

		key = I[:32]
		chainCode = I[32:]
	}

	return &ExtendedKey{
		Key:       append([]byte(nil), key...),
		ChainCode: append([]byte(nil), chainCode...),
	}, nil
}

func parseHardenedSegment(segment string) (uint32, error) {
	if !strings.HasSuffix(segment, "'") {
		return 0, fmt.Errorf("non-hardened segment is not supported for ed25519: %s", segment)
	}

	raw := strings.TrimSuffix(segment, "'")
	n, err := strconv.ParseUint(raw, 10, 31)
	if err != nil {
		return 0, fmt.Errorf("invalid path segment %q: %w", segment, err)
	}

	return uint32(n) + 0x80000000, nil
}

func hmacSHA512(key, data []byte) []byte {
	h := hmac.New(sha512.New, key)
	_, _ = h.Write(data)
	return h.Sum(nil)
}

func shouldTrackActivity(user *model.User) bool {
	now := time.Now().UTC()
	age := now.Sub(user.CreatedAt)

	if age >= 24*time.Hour && age <= 30*24*time.Hour {
		return true
	}
	return false
}

func loginMetadata(loginType string) string {
	b, _ := json.Marshal(map[string]string{"login_type": loginType})
	return string(b)
}
