//go:build integration

package v1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/go-clickhouse/ch"
	"go.uber.org/zap"

	"dd-prediction-api/config"
	handlerMiddleware "dd-prediction-api/internal/handler/middleware"
	"dd-prediction-api/internal/model"
	"dd-prediction-api/internal/service"
	"dd-prediction-api/internal/storage/cache"
	clickStorage "dd-prediction-api/internal/storage/click"
	"dd-prediction-api/internal/storage/repository"
	"dd-prediction-api/pkg/apikey"
	pkgauth "dd-prediction-api/pkg/jwt"
	"dd-prediction-api/pkg/mtype"
	pkgrepo "dd-prediction-api/pkg/repository"
)

const testJWTSecret = "integration-test-jwt-secret-key"

// ── Mocks ─────────────────────────────────────────────────────────────────────

type noopBlockedCache struct{}

func (n *noopBlockedCache) Block(_ context.Context, _ uuid.UUID) error   { return nil }
func (n *noopBlockedCache) Unblock(_ context.Context, _ uuid.UUID) error { return nil }
func (n *noopBlockedCache) IsBlocked(_ context.Context, _ uuid.UUID) (bool, error) {
	return false, nil
}

type noopDuelTxRepo struct{}

func (n *noopDuelTxRepo) BulkInsert(_ context.Context, _ []model.DuelTransaction) error {
	return nil
}
func (n *noopDuelTxRepo) GetByUserID(_ context.Context, _ uuid.UUID) ([]model.DuelTransaction, error) {
	return nil, nil
}

// ── Test application ──────────────────────────────────────────────────────────

type testApp struct {
	fiber    *fiber.App
	db       *bun.DB
	userRepo *repository.UserRepository
	projRepo *repository.ProjectRepository
	jwtAuth  pkgauth.JWTAuthenticator
}

func buildTestApp(t *testing.T) *testApp {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}

	// — Database ——————————————————————————————————————————————————————————————
	conf, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	require.NoError(t, err)
	db := bun.NewDB(stdlib.OpenDBFromPool(pool), pgdialect.New())
	require.NoError(t, db.Ping())
	t.Cleanup(func() { _ = db.Close() })

	// — Repositories ——————————————————————————————————————————————————————————
	userRepo := repository.NewUserRepository(
		pkgrepo.NewGenericRepository[model.User, uuid.UUID](db),
	)
	projRepo := repository.NewProjectRepository(
		pkgrepo.NewGenericRepository[model.Project, uuid.UUID](db),
	)
	duelRepo := repository.NewDuelRepository(
		pkgrepo.NewGenericRepository[model.Duel, uuid.UUID](db),
	)
	playerRepo := repository.NewPlayerRepository(
		pkgrepo.NewGenericRepository[model.Player, uuid.UUID](db),
	)
	projClaimRepo := repository.NewProjectCommissionClaimRepository(db)
	projAccrualRepo := repository.NewProjectCommissionAccrualRepository(db)
	txManager := pkgrepo.NewTransactionManager(db)

	// — Config ————————————————————————————————————————————————————————————————
	testConf := &config.Config{}
	testConf.Auth.SecretSignKey = testJWTSecret
	testConf.Auth.AccessTokenTTL = time.Hour
	testConf.Auth.RefreshTokenTTL = 24 * time.Hour
	testConf.App.DuelQueueSize = 10

	// — Services ——————————————————————————————————————————————————————————————
	jwtAuth := pkgauth.NewJWTAuth(testConf)
	jwtSvc := &service.JWTService{JWT: jwtAuth}

	userSvc := &service.UserService{
		UserRepository:     userRepo,
		TransactionManager: txManager,
	}

	projSvc := service.NewProjectService(
		testConf,
		projRepo,
		userRepo,
		nil,
		txManager,
		cache.NewBlockedProjectCache(nil),
		nil,
	)
	projSvc.BlockedProjectCache = &noopBlockedCache{}

	duelSvc := &service.DuelService{
		DuelRepository:   duelRepo,
		PlayerRepository: playerRepo,
		DuelTxRepository: &noopDuelTxRepo{},
		// WalletService and CoinService are nil: only needed for Solana mutations
	}

	duelQueue := service.NewDuelQueue(duelSvc, testConf)

	commSvc := &service.CommissionService{
		ProjectRepository:            projRepo,
		DuelRepository:               duelRepo,
		ProjectClaimRepo:             projClaimRepo,
		ProjectCommissionAccrualRepo: projAccrualRepo,
		// WalletService and CoinService nil: only needed for claim mutations
	}

	// ClickHouse audit log: connects lazily so Insert failures are silently
	// logged rather than breaking the handler response.
	chDB := ch.Connect(ch.WithAddr("localhost:19999"))
	auditLog := clickStorage.NewAdminAuditLogRepository(chDB)

	// — Handlers ——————————————————————————————————————————————————————————————
	authH := &AuthHandler{
		JWTService:     jwtSvc,
		UserService:    userSvc,
		ProjectService: projSvc,
		env:            "test",
	}
	userH := &UserHandler{
		UserService:    userSvc,
		DuelService:    duelSvc,
		ProjectService: projSvc,
		AuditLog:       auditLog,
	}
	projH := &ProjectHandler{
		ProjectService: projSvc,
		AuditLog:       auditLog,
	}
	duelH := &DuelHandler{
		DuelQueue:   duelQueue,
		DuelService: duelSvc,
	}
	adminH := &AdminHandler{
		DuelQueue:         duelQueue,
		DuelService:       duelSvc,
		CommissionService: commSvc,
		AuditLog:          auditLog,
	}
	commH := &CommissionHandler{
		CommissionService: commSvc,
	}

	// — Fiber app ——————————————————————————————————————————————————————————————
	app := fiber.New()
	logMW := handlerMiddleware.NewLoggingMiddleware(zap.NewNop())
	app.Use(logMW.LogErrorsDevelopment)

	authH.RegisterRoutes(app)
	userH.RegisterRoutes(app, authH)
	projH.RegisterRoutes(app, authH)
	duelH.RegisterRoutes(app, authH)
	adminH.RegisterRoutes(app, authH)
	commH.RegisterRoutes(app, authH)

	return &testApp{
		fiber:    app,
		db:       db,
		userRepo: userRepo,
		projRepo: projRepo,
		jwtAuth:  jwtAuth,
	}
}

// generateToken creates a signed Bearer token for the given user.
func (ta *testApp) generateToken(user *model.User) string {
	claims := pkgauth.TokenClaims{
		UserID:       user.ID,
		SessionID:    uuid.New(),
		Role:         user.Role,
		ProjectID:    user.ProjectID,
		RegisteredAt: user.CreatedAt,
	}
	pair, err := ta.jwtAuth.GenerateTokenPair(claims)
	if err != nil {
		panic("generateToken: " + err.Error())
	}
	return "Bearer " + pair.AccessToken.Token
}

// insertUser creates a user in the real DB and registers cleanup.
// Deleting the user cascades to any owned projects.
func (ta *testApp) insertUser(t *testing.T, role mtype.Role) *model.User {
	t.Helper()
	u := model.NewUser(role)
	require.NoError(t, ta.userRepo.Create(context.Background(), u))
	t.Cleanup(func() {
		_, _ = ta.db.NewDelete().Model((*model.User)(nil)).
			Where("id = ?", u.ID).Exec(context.Background())
	})
	return u
}

// insertProject creates a project with a properly signed API key and registers cleanup.
func (ta *testApp) insertProject(t *testing.T, partnerID uuid.UUID) *model.Project {
	t.Helper()
	projID := uuid.New()
	key := apikey.Generate(projID, testJWTSecret)
	p := model.NewProject(projID, key, partnerID)
	require.NoError(t, ta.projRepo.Create(context.Background(), p))
	t.Cleanup(func() {
		_, _ = ta.db.NewDelete().Model((*model.Project)(nil)).
			Where("id = ?", p.ID).Exec(context.Background())
	})
	return p
}

// insertPartnerWithProject creates a partner user, inserts a project for them,
// links the user's project_id, and returns both with an updated user.
func (ta *testApp) insertPartnerWithProject(t *testing.T) (*model.User, *model.Project) {
	t.Helper()
	partner := ta.insertUser(t, mtype.RolePartner)
	proj := ta.insertProject(t, partner.ID)

	_, err := ta.db.NewUpdate().Model((*model.User)(nil)).
		Set("project_id = ?", proj.ID).
		Where("id = ?", partner.ID).
		Exec(context.Background())
	require.NoError(t, err)
	partner.ProjectID = proj.ID

	return partner, proj
}

// ── POST /auth/refresh ────────────────────────────────────────────────────────

func TestRefreshTokens_noToken_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRefreshTokens_invalidToken_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", nil)
	req.Header.Set("Authorization", "Bearer garbage.token.value")
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ── POST /auth/send-code ──────────────────────────────────────────────────────

func TestSendCode_invalidEmail_returns400(t *testing.T) {
	ta := buildTestApp(t)

	body := strings.NewReader(`{"email":"not-a-valid-email"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/send-code", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSendCode_malformedBody_returns400(t *testing.T) {
	ta := buildTestApp(t)

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/auth/send-code", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ── POST /auth/sign-in-email ──────────────────────────────────────────────────

func TestSignInEmail_invalidEmail_returns400(t *testing.T) {
	ta := buildTestApp(t)

	body := strings.NewReader(`{"email":"not-valid","code":"123456"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in-email", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ── POST /auth/sign-in-google ─────────────────────────────────────────────────

// authWithGoogle returns 401 before touching Firebase when no Authorization header is present.
func TestSignInGoogle_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in-google", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ── GET /user/me ──────────────────────────────────────────────────────────────

func TestGetMe_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetMe_invalidToken_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.valid")
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Admin JWT bypasses the X-API-Key requirement, so GET /user/me works with
// just an Authorization header when the role is Admin.
func TestGetMe_adminJWT_returnsOwnProfile(t *testing.T) {
	ta := buildTestApp(t)

	user := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(user)

	req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got model.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, user.ID, got.ID)
	assert.Equal(t, mtype.Role(mtype.RoleAdmin), got.Role)
	assert.True(t, got.IsActive)
}

// A partner sends their project's API key together with the JWT.
// The API key middleware validates the key and checks that the JWT's ProjectID
// matches the one embedded in the key.
func TestGetMe_partnerWithAPIKey_returnsOwnProfile(t *testing.T) {
	ta := buildTestApp(t)

	partner, proj := ta.insertPartnerWithProject(t)
	token := ta.generateToken(partner)

	req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
	req.Header.Set("Authorization", token)
	req.Header.Set("X-API-Key", proj.APIKey)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got model.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, partner.ID, got.ID)
}

// ── GET /user/me/financial-history ───────────────────────────────────────────

func TestGetFinancialHistory_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/user/me/financial-history", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Admin JWT bypasses API key; noopDuelTxRepo returns an empty list.
func TestGetFinancialHistory_adminJWT_returnsEmptyList(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/user/me/financial-history", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got []model.FinancialTxEntry
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Empty(t, got)
}

// ── GET /user/:id (admin) ─────────────────────────────────────────────────────

func TestGetUserByID_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/user/"+uuid.New().String(), nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetUserByID_nonAdmin_returns403(t *testing.T) {
	ta := buildTestApp(t)

	partner, proj := ta.insertPartnerWithProject(t)
	token := ta.generateToken(partner)
	target := ta.insertUser(t, mtype.RoleUser)

	req := httptest.NewRequest(http.MethodGet, "/user/"+target.ID.String(), nil)
	req.Header.Set("Authorization", token)
	req.Header.Set("X-API-Key", proj.APIKey)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetUserByID_admin_returnsUser(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	target := ta.insertUser(t, mtype.RoleUser)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/user/"+target.ID.String(), nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got model.User
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, target.ID, got.ID)
}

func TestGetUserByID_invalidID_returns400(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/user/not-a-uuid", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ── POST /project-handler/ ────────────────────────────────────────────────────

func TestCreateProject_nonPartner_returns401(t *testing.T) {
	ta := buildTestApp(t)

	// CheckForPartnerPermissions returns 401 for non-partner roles.
	user := ta.insertUser(t, mtype.RoleUser)
	token := ta.generateToken(user)

	req := httptest.NewRequest(http.MethodPost, "/project-handler/", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateProject_noToken_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/project-handler/", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateProject_partner_createsProjectAndReturns200(t *testing.T) {
	ta := buildTestApp(t)

	partner := ta.insertUser(t, mtype.RolePartner)
	token := ta.generateToken(partner)

	req := httptest.NewRequest(http.MethodPost, "/project-handler/", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var proj model.Project
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&proj))
	assert.Equal(t, partner.ID, proj.PartnerID)
	assert.NotEmpty(t, proj.APIKey)
	assert.NotEqual(t, uuid.Nil, proj.ID)

	found, err := ta.projRepo.GetByID(context.Background(), proj.ID)
	require.NoError(t, err)
	assert.Equal(t, proj.ID, found.ID)

	t.Cleanup(func() {
		_, _ = ta.db.NewDelete().Model((*model.Project)(nil)).
			Where("id = ?", proj.ID).Exec(context.Background())
	})
}

func TestCreateProject_partner_alreadyHasProject_returns409(t *testing.T) {
	ta := buildTestApp(t)

	partner := ta.insertUser(t, mtype.RolePartner)
	_ = ta.insertProject(t, partner.ID)
	token := ta.generateToken(partner)

	req := httptest.NewRequest(http.MethodPost, "/project-handler/", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

// ── PUT /project/:id/block-key & /unblock-key ─────────────────────────────────

func TestBlockKey_nonAdmin_returns403(t *testing.T) {
	ta := buildTestApp(t)

	partner, proj := ta.insertPartnerWithProject(t)
	token := ta.generateToken(partner)

	req := httptest.NewRequest(http.MethodPut, "/project/"+proj.ID.String()+"/block-key", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestBlockKey_admin_setsBlockedTrueInDB(t *testing.T) {
	ta := buildTestApp(t)

	partner := ta.insertUser(t, mtype.RolePartner)
	proj := ta.insertProject(t, partner.ID)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPut, "/project/"+proj.ID.String()+"/block-key", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	blocked, err := ta.projRepo.IsBlocked(context.Background(), proj.ID)
	require.NoError(t, err)
	assert.True(t, blocked, "project should be marked blocked in DB")
}

func TestUnblockKey_admin_setsBlockedFalseInDB(t *testing.T) {
	ta := buildTestApp(t)

	partner := ta.insertUser(t, mtype.RolePartner)
	proj := ta.insertProject(t, partner.ID)

	require.NoError(t, ta.projRepo.SetBlocked(context.Background(), proj.ID, true))

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPut, "/project/"+proj.ID.String()+"/unblock-key", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	blocked, err := ta.projRepo.IsBlocked(context.Background(), proj.ID)
	require.NoError(t, err)
	assert.False(t, blocked, "project should be unblocked in DB")
}

func TestBlockKey_invalidProjectID_returns400(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPut, "/project/not-a-uuid/block-key", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ── GET /duel/* ───────────────────────────────────────────────────────────────

// All /duel routes require ApiKeyMiddleware. A request with no credentials
// hits the "API key is required" guard before any handler logic runs.
func TestGetAllDuels_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/duel/all", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Admin JWT bypasses the API key requirement. An empty database returns [].
func TestGetAllDuels_adminJWT_returnsEmptyList(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/duel/all", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got []model.DuelShow
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Empty(t, got)
}

func TestGetMyDuels_adminJWT_returnsEmptyList(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/duel/my", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetAllDuelsWithJoined_adminJWT_returnsEmptyList(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/duel/all-with-joined", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetParticipatingDuels_adminJWT_returnsEmptyList(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/duel/participating", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestGetDuelHistory_adminJWT_returnsEmptyList(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/duel/history", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

// GET /duel/token-accounts calls WalletService which is nil in tests.
// Only the no-auth path is safe to test.
func TestGetTokenAccounts_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/duel/token-accounts", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ── Duel mutation routes (no-auth only — Solana not available in tests) ───────

func TestCreateDuelSolana_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/duel/solana/", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSignCreateDuelTx_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/duel/solana/sign-tx", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestJoinDuelSolana_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/duel/solana/join", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSignJoinDuelTx_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/duel/solana/join/sign-tx", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSelfResolveDuel_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPut, "/duel/self-resolve", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ── GET /me/dashboard (partner commission dashboard) ─────────────────────────

func TestGetDashboard_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/me/dashboard", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Partner with a valid API key and a JWT whose ProjectID matches the project
// gets their dashboard (empty stats on a fresh database).
func TestGetDashboard_partnerWithAPIKey_returnsData(t *testing.T) {
	ta := buildTestApp(t)

	partner, proj := ta.insertPartnerWithProject(t)
	token := ta.generateToken(partner)

	req := httptest.NewRequest(http.MethodGet, "/me/dashboard", nil)
	req.Header.Set("Authorization", token)
	req.Header.Set("X-API-Key", proj.APIKey)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got model.PartnerDashboard
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, proj.APIKey, got.APIKey)
	assert.False(t, got.IsBlocked)
}

// ── POST /commission/claim & GET /commission/history (middleware bug) ─────────

// BUG: CommissionHandler.RegisterRoutes creates the /commission group at the
// application root without middleware. c.Locals("claims") is never set, so
// ClaimProject always returns 401 regardless of credentials provided.
func TestClaimProjectCommission_alwaysReturns401(t *testing.T) {
	ta := buildTestApp(t)

	body := strings.NewReader(`{"wallet_address":"someaddr"}`)
	req := httptest.NewRequest(http.MethodPost, "/commission/claim", body)
	req.Header.Set("Content-Type", "application/json")
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// BUG: same middleware omission — GetProjectHistory always returns 401.
func TestGetProjectCommissionHistory_alwaysReturns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/commission/history", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// ── /admin/* routes ───────────────────────────────────────────────────────────

// BUG: CheckForAdminRights has an inverted condition:
//   if claims.Role.HasProjectAdminRights() { return Forbidden }
// HasProjectAdminRights() returns true for Admin (3), Partner (2), PartnerAdmin (1).
// Only RoleUser (0) passes — making all /admin/* routes inaccessible to admins.

func TestAdminCreateDuel_noAuth_returns401(t *testing.T) {
	ta := buildTestApp(t)

	req := httptest.NewRequest(http.MethodPost, "/admin/duel/", nil)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// Admin JWT → 403 due to inverted CheckForAdminRights check.
func TestAdminCreateDuel_adminJWT_returns403(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPost, "/admin/duel/", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminApproveDuel_adminJWT_returns403(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPut, "/admin/duel/approve", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminResolveDuel_adminJWT_returns403(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPut, "/admin/duel/resolve", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminCancelDuel_adminJWT_returns403(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPut, "/admin/duel/cancel", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminEditDuel_adminJWT_returns403(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPut, "/admin/duel/edit", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminClaimDDProfit_adminJWT_returns403(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodPost, "/admin/commission/dd-profit/claim", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminGetDDProfitHistory_adminJWT_returns403(t *testing.T) {
	ta := buildTestApp(t)

	admin := ta.insertUser(t, mtype.RoleAdmin)
	token := ta.generateToken(admin)

	req := httptest.NewRequest(http.MethodGet, "/admin/commission/dd-profit/history", nil)
	req.Header.Set("Authorization", token)
	resp, err := ta.fiber.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
