package auth

import (
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/mtype"

	"github.com/google/uuid"
)

var (
	ErrGenerateToken        = "failed to generate token"
	ErrInvalidSigningMethod = "unexpected signing method"
	ErrInvalidToken         = "token is not valid"
	ErrInvalidTokenClaims   = "invalid token claims"
	ErrInvalidUserIDClaim   = "invalid user_id claim"
)

type JWTAuthenticator interface {
	GenerateTokenPair(options TokenClaims) (*TokenPairWithExpiration, error)
	ParseToken(authToken string) (*TokenClaims, error)
}

type TokenPairWithExpiration struct {
	AccessToken  *TokenExpiration
	RefreshToken *TokenExpiration
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type TokenClaims struct {
	UserID       uuid.UUID  `json:"user_id"`
	SessionID    uuid.UUID  `json:"session_id"`
	ProjectID    uuid.UUID  `json:"project_id"`
	Role         mtype.Role `json:"role"`
	RegisteredAt time.Time  `json:"registered_at"`
}

func (c *TokenClaims) RefreshSessionID() bool {
	if c == nil {
		return false
	}

	sessionID, err := uuid.NewV7()
	if err != nil {
		return false
	}

	c.SessionID = sessionID
	return true
}

func NewTokenClaimsByUser(
	user *model.User,
) (TokenClaims, bool) {
	sessionID, err := uuid.NewV7()
	if err != nil {
		return TokenClaims{}, false
	}

	return TokenClaims{
		UserID:       user.ID,
		SessionID:    sessionID,
		Role:         user.Role,
		ProjectID:    user.ProjectID,
		RegisteredAt: user.CreatedAt,
	}, true
}
