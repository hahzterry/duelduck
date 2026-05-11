package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"dd-prediction-api/config"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/service"
	"dd-prediction-api/pkg/apperrors"
	"dd-prediction-api/pkg/captcha"
	auth "dd-prediction-api/pkg/jwt"
	"dd-prediction-api/pkg/mtype"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

const (
	jwtAuthType       = "Bearer"
	ratePeriod  int64 = 1 // seconds
)

type RateLimit struct {
	CountForPeriod uint32
	LastRequest    int64
}

type uploadBucket struct {
	minute int64
	count  int
}

var (
	userImageUploadCache = make(map[uuid.UUID]uploadBucket)
	uploadImageMu        sync.Mutex
)

var (
	RateLimiter   = make(map[string]*RateLimit)
	RateLimiterMu sync.Mutex
	rateLimit     uint32 = 6
)

type AuthHandler struct {
	Firebase         *firebaseAuth.Client
	JWTService       *service.JWTService
	AuthService      *service.AuthService
	UserService      *service.UserService
	ProjectService   *service.ProjectService
	TurnstileService *captcha.TurnstileService
	env              string
}

func NewAuthHandler(
	c *config.Config,
	userService *service.UserService,
	authService *service.AuthService,
	jwtService *service.JWTService,
	projectService *service.ProjectService,
) (*AuthHandler, error) {
	authHandler := &AuthHandler{
		UserService:      userService,
		JWTService:       jwtService,
		AuthService:      authService,
		ProjectService:   projectService,
		TurnstileService: captcha.NewTurnstileService(c.Auth.TurnstileSecret),
		env:              c.App.Environment,
	}

	// exceed rate limit for stage
	if authHandler.env == "stage" {
		rateLimit = 100
	}

	err := authHandler.setFirebaseAuth(c.App.FirebaseFilePath)
	if err != nil {
		return nil, err
	}

	return authHandler, nil
}

func (h *AuthHandler) setFirebaseAuth(firebaseFilePath string) error {
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, firebaseFilePath)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("firebase.NewApp: %w", err)
	}

	h.Firebase, err = app.Auth(context.Background())
	if err != nil {
		return fmt.Errorf("app.Auth: %w", err)
	}

	return nil
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	authGroup := app.Group("/auth")
	{
		authGroup.Post("/refresh", h.RefreshTokens)

		protected := authGroup.Group("/")
		if h.env == "prod" {
			protected.Use(h.TurnstileMiddleware)
		}

		protected.Post("/send-code", h.SendCode)
		protected.Post("/sign-in-email", h.SignInWithEmail)
		protected.Post("/sign-in-google", h.SignInWithGoogleMiddleware, h.SignInWithGoogle)
	}
}

func (h *AuthHandler) SignInWithGoogleMiddleware(c fiber.Ctx) error {
	if err := h.authWithGoogle(c); err != nil {
		return err
	}

	return c.Next()
}

func (h *AuthHandler) TurnstileMiddleware(c fiber.Ctx) error {
	if h.TurnstileService == nil {
		zap.L().Warn("Turnstile verification is not set, skipping")
		return c.Next()
	}

	var token string

	token = c.Get("Cf-Turnstile-Response")
	if token == "" {
		bodyBytes := c.Body()
		if len(bodyBytes) > 0 {
			var body map[string]interface{}
			if err := json.Unmarshal(bodyBytes, &body); err == nil {
				if val, ok := body["cf-turnstile-response"].(string); ok && val != "" {
					token = val
				}
			}
		}
	}

	if token == "" {
		return apperrors.Forbidden("verification is not provided")
	}

	remoteIP := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			remoteIP = strings.TrimSpace(ips[0])
		}
	}

	verifyResp, err := h.TurnstileService.VerifyToken(token, remoteIP)
	if err != nil {
		zap.L().Error("Turnstile verification error", zap.Error(err))
		return apperrors.Internal("verification error")
	}

	if !verifyResp.Success {
		zap.L().Warn("Turnstile verification failed", zap.Strings("error_codes", verifyResp.ErrorCodes))
		return apperrors.Forbidden("bot detected - verification failed")
	}

	return c.Next()
}

func (h *AuthHandler) authWithGoogle(c fiber.Ctx) error {
	strToken := c.Get("Authorization")
	if strToken == "" {
		return apperrors.Unauthorized("authorization token not found")
	}
	strToken = strings.TrimPrefix(strToken, jwtAuthType+" ")

	token, err := h.Firebase.VerifyIDTokenAndCheckRevoked(c.Context(), strToken)
	if err != nil {
		if firebaseAuth.IsIDTokenExpired(err) {
			return apperrors.Unauthorized("authorization token is expired")
		}

		return apperrors.Unauthorized("invalid authorization token", err)
	}

	emailRaw, ok := token.Claims["email"].(string)
	if !ok {
		return apperrors.Unauthorized("google token: email not found")
	}

	email, ok := mtype.NewEmail(emailRaw)
	if !ok {
		return apperrors.Unauthorized("google token: email is invalid")
	}
	c.Locals("email", email)

	if pictureRaw, ok := token.Claims["picture"].(string); ok {
		c.Locals("google_image_url", strings.TrimSpace(pictureRaw))
	}

	return nil
}

// SignInWithGoogle godoc
//
//	@Summary		Sign in user with Google authentication
//	@Description	Authenticates a user using Google credentials and returns user data with JWT tokens. The email is validated through Google's authentication system and a new session is created.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		GoogleAuth
//	@Success		200	{object}	object{user=model.User,jwt_info=auth.TokenPair,daily_reward=int}	"Successfully authenticated"
//	@Failure		400	{object}	apperrors.ErrorPublic												"Invalid request data or missing email"
//	@Failure		401	{object}	apperrors.ErrorPublic												"Unauthorized - Invalid Google token"
//	@Failure		404	{object}	apperrors.ErrorPublic												"User not found"
//	@Failure		500	{object}	apperrors.ErrorPublic												"Internal server error"
//	@Router			/auth/sign-in-google [post]
func (h *AuthHandler) SignInWithGoogle(c fiber.Ctx) error {
	email, ok := c.Locals("email").(mtype.Email)
	if !ok {
		return apperrors.BadRequest("email is missing")
	}
	googleImageURL, _ := c.Locals("google_image_url").(string)

	user, err := h.UserService.SignInWithGoogle(
		c.Context(),
		email,

		googleImageURL,
	)
	if err != nil {
		return err
	}

	claims, ok := auth.NewTokenClaimsByUser(user)
	if !ok {
		return apperrors.Internal("failed to generate token claims for user")
	}

	tokenPair, err := h.JWTService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"jwt_info": tokenPair,
	})
}

// RefreshTokens godoc
//
//	@Summary		Refresh JWT tokens
//	@Description	Refreshes the user's JWT tokens using the provided refresh token from the Authorization header. Returns a new pair of access and refresh tokens.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string							true	"Bearer refresh token"	default(Bearer <refresh_token>)
//	@Success		200				{object}	object{jwt_info=auth.TokenPair}	"Successfully refreshed tokens"
//	@Failure		401				{object}	apperrors.ErrorPublic			"Unauthorized - Invalid or expired refresh token"
//	@Failure		500				{object}	apperrors.ErrorPublic			"Internal server error"
//	@Router			/auth/refresh [post]
func (h *AuthHandler) RefreshTokens(c fiber.Ctx) error {
	token := c.Get("Authorization")

	tokenPair, err := h.JWTService.RefreshSession(c.Context(), token)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"jwt_info": tokenPair,
	})
}

// SendCode godoc
//
//	@Summary		Send verification code to email
//	@Description	Sends a verification code to the specified email address. The code is required for email-based sign-in or sign-up.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.SendCode			true	"Email address to send the verification code to"
//	@Success		200		{object}	nil						"Verification code sent successfully"
//	@Failure		400		{object}	apperrors.ErrorPublic	"Invalid request data or email format"
//	@Failure		500		{object}	apperrors.ErrorPublic	"Internal server error during code sending"
//	@Router			/auth/send-code [post]
func (h *AuthHandler) SendCode(c fiber.Ctx) error {
	var req model.SendCode
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	email, ok := mtype.NewEmail(req.Email)
	if !ok {
		return apperrors.BadRequest("invalid email")
	}

	err := h.AuthService.SendCodeOnEmail(c.Context(), "index.html", email)
	if err != nil {
		return err
	}

	return nil
}

// SignInWithEmail godoc
//
//	@Summary		User Sign-In/Sign-Up with Email
//	@Description	Authenticates a user or creates a new account using an email and a verification code.
//	@Description	A verification code must be obtained first via the /send-code endpoint.
//	@Description	If the user does not exist, a new account is created. An optional referrer token can be provided to link the new user to a referrer.
//	@Description	On successful authentication, it returns the user's profile, a JWT token pair (access and refresh), and information about their daily login streak reward.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.SignInWithEmail							true	"Sign-in request with email, verification code, and optional referrer token"
//	@Success		200		{object}	object{user=model.User,jwt_info=auth.TokenPair}	"Successful sign-in or sign-up, returning user data, JWT tokens, and daily reward info"
//	@Failure		400		{object}	apperrors.ErrorPublic							"Invalid request data or email format"
//	@Failure		401		{object}	apperrors.ErrorPublic							"Invalid verification code"
//	@Failure		404		{object}	apperrors.ErrorPublic							"User not found (error during sign-in process)"
//	@Failure		500		{object}	apperrors.ErrorPublic							"Internal server error during token generation or other processes"
//	@Router			/auth/sign-in-email [post]
func (h *AuthHandler) SignInWithEmail(c fiber.Ctx) error {
	var req model.SignInWithEmail
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	email, ok := mtype.NewEmail(req.Email)
	if !ok {
		return apperrors.BadRequest("invalid email")
	}

	ok, err := h.AuthService.CheckUsersEmailCode(c.Context(), email, req.Code)
	if err != nil {
		return err
	}

	if !ok {
		return apperrors.Unauthorized("invalid code")
	}

	user, err := h.UserService.SignInWithEmail(c.Context(), email)
	if err != nil {
		return apperrors.NotFound("user with email not found", err)
	}

	claims, ok := auth.NewTokenClaimsByUser(user)
	if !ok {
		return apperrors.Internal("failed to generate token claims for user")
	}

	tokenPair, err := h.JWTService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"jwt_info": tokenPair,
	})
}

// SignInWithWallet godoc
//
//	@Summary		Sign in user with wallet
//	@Description	Authenticates a user using a connected crypto wallet (e.g., Solana Phantom). Accepts wallet-related data and returns user data with JWT tokens.
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			request	body		model.AuthWithWallet												true	"Wallet sign in request"
//	@Success		200		{object}	object{user=model.User,jwt_info=auth.TokenPair,daily_reward=int}	"Successfully authenticated"
//	@Failure		400		{object}	apperrors.ErrorPublic												"Invalid request data"
//	@Failure		401		{object}	apperrors.ErrorPublic												"Unauthorized - Invalid wallet signature or data"
//	@Failure		403		{object}	apperrors.ErrorPublic												"Forbidden - Bot detected - verification failed"
//	@Failure		404		{object}	apperrors.ErrorPublic												"User not found"
//	@Failure		500		{object}	apperrors.ErrorPublic												"Internal server error"
//	@Router			/auth/sign-in-wallet [post]
func (h *AuthHandler) SignInWithWallet(c fiber.Ctx) error {
	var req model.AuthWithWallet
	if err := c.Bind().JSON(&req); err != nil {
		return apperrors.BadRequest("invalid request data")
	}

	user, err := h.UserService.SignInWithWallet(c.Context(), req)
	if err != nil {
		return err
	}

	claims, ok := auth.NewTokenClaimsByUser(user)
	if !ok {
		return apperrors.Internal("failed to generate token claims for user")
	}

	tokenPair, err := h.JWTService.GenerateTokenPair(c.Context(), claims)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"user":     user,
		"jwt_info": tokenPair,
	})
}

func (h *AuthHandler) AuthMiddleware(c fiber.Ctx) error {
	token := c.Get("Authorization")

	claims, err := h.JWTService.ParseToken(token)
	if err != nil {
		return apperrors.Unauthorized("failed to parse token", err)
	}

	if claims == nil {
		return apperrors.Internal("failed to generate claims")
	}

	c.Locals("claims", *claims)

	track := false
	if !claims.RegisteredAt.IsZero() {
		now := time.Now().UTC()
		age := now.Sub(claims.RegisteredAt)

		if age >= 24*time.Hour && age <= 30*24*time.Hour {
			track = true
		}
	}

	ctx := context.WithValue(
		c.Context(),
		model.CtxTrackActivityKey,
		track,
	)

	c.SetContext(ctx)

	return c.Next()
}

func (h *AuthHandler) AuthMiddlewareQuery(c fiber.Ctx) error {
	token := c.Query("token")

	claims, err := h.JWTService.ParseToken(token)
	if err != nil {
		return apperrors.Unauthorized("failed to parse token", err)
	}

	if claims == nil {
		return apperrors.Internal("failed to generate claims")
	}

	c.Locals("claims", *claims)

	return c.Next()
}

// ApiKeyMiddleware enforces API key authentication for project-scoped routes.
// Global admins (RoleAdmin) may bypass the key by providing a valid JWT instead.
// The project ID is decoded directly from the key — no database lookup required.
// An API key from project A cannot be used to access project B's resources.
func (h *AuthHandler) ApiKeyMiddleware(c fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")

	if apiKey == "" {
		token := c.Get("Authorization")
		claims, err := h.JWTService.ParseToken(token)
		if err != nil || claims == nil || !claims.Role.Admin() {
			return apperrors.Unauthorized("API key is required")
		}
		c.Locals("claims", *claims)
		return c.Next()
	}

	projectID, ok := h.ProjectService.ParseAPIKey(apiKey)
	if !ok {
		return apperrors.Unauthorized("invalid API key")
	}

	blocked, err := h.ProjectService.IsAPIKeyBlocked(c.Context(), projectID)
	if err != nil {
		return apperrors.Internal("failed to check API key status")
	}
	if blocked {
		return apperrors.Forbidden("API key is blocked")
	}

	// Project isolation: JWT user must belong to the same project.
	token := c.Get("Authorization")
	if token != "" {
		claims, err := h.JWTService.ParseToken(token)
		if err == nil && claims != nil && !claims.Role.Admin() {
			if claims.ProjectID != projectID {
				return apperrors.Forbidden("API key does not belong to your project")
			}
			c.Locals("claims", *claims)
		}
	}

	c.Locals("project_id", projectID)

	return c.Next()
}

func (h *AuthHandler) CheckForAdminPermissions(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.Role.Admin() {
		return apperrors.Forbidden("admin permission required")
	}

	return c.Next()
}

func (h *AuthHandler) CheckForAdminRights(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.Role.HasProjectAdminRights() {
		return apperrors.Forbidden("admin rights required")
	}

	return c.Next()
}

func (h *AuthHandler) RequireUserRole(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	if !claims.Role.User() {
		return apperrors.Forbidden("only regular users can perform this action")
	}

	return c.Next()
}

func (h *AuthHandler) CheckForPartnerPermissions(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	user, err := h.UserService.GetByID(c.Context(), claims.UserID)
	if err != nil {
		return apperrors.Unauthorized("failed to fetch user")
	}

	if !user.Role.Partner() {
		return apperrors.Unauthorized("permission denied")
	}

	return c.Next()
}

func (h *AuthHandler) UploadUserImageMiddleware(c fiber.Ctx) error {
	claims, ok := c.Locals("claims").(auth.TokenClaims)
	if !ok {
		return apperrors.Unauthorized("claims not found")
	}

	currentMinute := time.Now().Truncate(time.Minute).Unix()

	uploadImageMu.Lock()
	defer uploadImageMu.Unlock()

	b := userImageUploadCache[claims.UserID]

	if b.minute != currentMinute {
		b.minute = currentMinute
		b.count = 0
	}

	if b.count >= 5 {
		return apperrors.TooManyRequests("only 5 uploads per minute allowed")
	}

	b.count++
	userImageUploadCache[claims.UserID] = b

	return c.Next()
}
