package service

import (
	"context"
	"strings"
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/storage/cache"
	"dd-prediction-api/internal/storage/click"
	"dd-prediction-api/pkg/apperrors"
	auth "dd-prediction-api/pkg/jwt"

	"github.com/google/uuid"
)

type JWTService struct {
	Storage *cache.JWTStorage
	JWT     auth.JWTAuthenticator
}

func NewJWTService(storage *cache.JWTStorage, jwt auth.JWTAuthenticator) *JWTService {
	return &JWTService{
		Storage: storage,
		JWT:     jwt,
	}
}

func (s *JWTService) GenerateTokenPair(ctx context.Context, claims auth.TokenClaims) (*auth.TokenPair, error) {
	tokenPair, err := s.JWT.GenerateTokenPair(claims)
	if err != nil {
		return nil, err
	}

	err = s.Storage.Save(ctx, tokenPair.RefreshToken, claims)
	if err != nil {
		return nil, err
	}

	return &auth.TokenPair{
		AccessToken:  tokenPair.AccessToken.Token,
		RefreshToken: tokenPair.RefreshToken.Token,
	}, nil
}

func (s *JWTService) RefreshTokenPair(
	ctx context.Context,
	claims auth.TokenClaims,
	prevSessionID uuid.UUID,
) (*auth.TokenPair, error) {
	err := s.Storage.DeleteUserSession(ctx, claims.UserID, prevSessionID)
	if err != nil {
		return nil, err
	}

	tokenPair, err := s.JWT.GenerateTokenPair(claims)
	if err != nil {
		return nil, err
	}

	err = s.Storage.Save(ctx, tokenPair.RefreshToken, claims)
	if err != nil {
		return nil, err
	}

	return &auth.TokenPair{
		AccessToken:  tokenPair.AccessToken.Token,
		RefreshToken: tokenPair.RefreshToken.Token,
	}, nil
}

func (s *JWTService) GetRefreshByUserID(
	ctx context.Context,
	userID uuid.UUID,
	sessionID uuid.UUID,
) (string, error) {
	refreshToken, err := s.Storage.GetUserSession(ctx, userID, sessionID)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *JWTService) RefreshSession(
	ctx context.Context,
	token string,
) (*auth.TokenPair, error) {
	claims, err := s.ParseToken(token)
	if err != nil {
		return nil, err
	}

	if claims == nil {
		return nil, apperrors.Unauthorized("invalid token claims")
	}

	storedRefresh, err := s.GetRefreshByUserID(ctx, claims.UserID, claims.SessionID)
	if err != nil {
		return nil, err
	}

	token = strings.TrimPrefix(token, "Bearer ")
	if token != storedRefresh {
		return nil, apperrors.Unauthorized("refresh token does not match with stored one")
	}

	prevSessionID := claims.SessionID
	claims.RefreshSessionID()

	tokenPair, err := s.RefreshTokenPair(ctx, *claims, prevSessionID)
	if err != nil {
		return nil, err
	}

	if model.ShouldTrackActivity(ctx) {
		if buf := click.GetActivityBuffer(); buf != nil {
			buf.Add(model.UserActivity{
				UserID:        claims.UserID,
				ActivityType:  model.ActivityTokenRefresh,
				CreatedAt:     time.Now().UTC(),
				ActivityCount: 1,
			})
		}
	}

	return tokenPair, nil
}

func (s *JWTService) ParseToken(token string) (*auth.TokenClaims, error) {
	return s.JWT.ParseToken(token)
}

func (s *JWTService) ReissueSameSession(ctx context.Context, claims auth.TokenClaims) (*auth.TokenPair, error) {
	p, err := s.JWT.GenerateTokenPair(claims)
	if err != nil {
		return nil, err
	}
	if err = s.Storage.Save(ctx, p.RefreshToken, claims); err != nil {
		return nil, err
	}
	return &auth.TokenPair{
		AccessToken:  p.AccessToken.Token,
		RefreshToken: p.RefreshToken.Token,
	}, nil
}
