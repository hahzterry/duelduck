package service

import (
	"context"
	cryptoRand "crypto/rand"
	"dd-prediction-api/config"
	"dd-prediction-api/internal/storage/cache"
	"dd-prediction-api/pkg/apperrors"
	"dd-prediction-api/pkg/mailer"
	"dd-prediction-api/pkg/mtype"
	"fmt"
	"math"
	"math/big"
	"strings"

	emailverifier "github.com/AfterShip/email-verifier"
)

const (
	EmailVerificationCodeTopic = "DuelDuck: Verification Code"
)

type AuthService struct {
	CodeStorage   *cache.CodeStorage
	Mailer        mailer.Mailer
	EmailVerifier *emailverifier.Verifier
}

func NewAuthService(
	c *config.Config,
	codeStorage *cache.CodeStorage,
	mailer mailer.Mailer,
) (*AuthService, error) {
	emailVerifier := emailverifier.NewVerifier().
		EnableSMTPCheck().
		EnableCatchAllCheck().
		EnableDomainSuggest().
		EnableAutoUpdateDisposable().
		FromEmail(c.Brevo.BrevoEmail).
		HelloName(domainFromEmail(c.Brevo.BrevoEmail, c.Brevo.BrevoName))

	_ = emailVerifier.EnableAPIVerifier(emailverifier.YAHOO)

	return &AuthService{
		CodeStorage:   codeStorage,
		Mailer:        mailer,
		EmailVerifier: emailVerifier,
	}, nil
}

func (s *AuthService) SendCodeOnEmail(ctx context.Context, template string, email mtype.Email) error {
	valid, disposable, err := s.VerifyEmailForSignIn(email)
	if err != nil {
		return apperrors.ServiceUnavailable("failed to verify email", err)
	}
	if !valid {
		if disposable {
			return apperrors.BadRequest("disposable email addresses are not allowed")
		}
		return apperrors.BadRequest("invalid or undeliverable email")
	}

	code, err := generateVerificationCode()
	if err != nil {
		return err
	}

	code = strings.TrimSpace(code)
	_ = s.CodeStorage.Delete(ctx, email.String())
	err = s.CodeStorage.Save(ctx, email, code)
	if err != nil {
		return err
	}

	mail := mailer.NewMail(email, template, EmailVerificationCodeTopic, code)

	err = s.Mailer.SendEmail(ctx, mail)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) VerifyEmailForSignIn(email mtype.Email) (bool, bool, error) {
	if s.EmailVerifier == nil {
		return false, false, apperrors.Internal("email verifier is not initialized")
	}

	result, err := s.EmailVerifier.Verify(email.String())
	if err != nil {
		return false, false, err
	}

	if !result.Syntax.Valid {
		return false, false, nil
	}

	if result.Disposable {
		return false, true, nil
	}

	if !result.HasMxRecords {
		return false, false, nil
	}

	if result.SMTP != nil && !result.SMTP.Deliverable {
		return false, false, nil
	}

	return true, false, nil
}

func domainFromEmail(email string, fallback string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] == "" {
		if fallback == "" {
			return "localhost"
		}

		return fallback
	}

	return parts[1]
}

func (s *AuthService) CheckUsersEmailCode(ctx context.Context, email mtype.Email, code string) (bool, error) {
	storedCode, err := s.CodeStorage.GetCodeByEmail(ctx, email)
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(storedCode) != strings.TrimSpace(code) {
		return false, nil
	}

	return true, nil
}

func (s *AuthService) DeleteEmailCode(ctx context.Context, key string) error {
	return s.CodeStorage.Delete(ctx, key)
}

func (s *AuthService) SendCodeForEmailChange(ctx context.Context, template string, email mtype.Email) error {
	code, err := generateVerificationCode()
	if err != nil {
		return err
	}

	key := email.String() + "_emailChange"
	err = s.CodeStorage.SaveWithKey(ctx, key, code)
	if err != nil {
		return err
	}

	mail := mailer.NewMail(email, template, EmailVerificationCodeTopic, code)

	err = s.Mailer.SendEmail(ctx, mail)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) CheckEmailChangeCode(ctx context.Context, email mtype.Email, code string) (bool, error) {
	key := email.String() + "_emailChange"
	storedCode, err := s.CodeStorage.GetCodeByKey(ctx, key)
	if err != nil {
		return false, err
	}

	if storedCode != code {
		return false, nil
	}

	return true, nil
}

func generateRandomNum() (uint64, error) {
	randomNum, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(math.MaxUint32))
	if err != nil {
		return 0, apperrors.Internal("failed to generate num", err)
	}

	return randomNum.Uint64(), nil
}

func generateVerificationCode() (string, error) {
	num, err := generateRandomNum()
	if err != nil {
		return "", err
	}

	code := int(num)%900000 + 100000

	return fmt.Sprintf("%06d", code), nil
}
