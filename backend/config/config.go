package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

func NewConfig() *Config {
	config := env.Must[Config](env.ParseAs[Config]())

	return &config
}

const (
	EnvironmentProduction  = "prod"
	EnvironmentStage       = "stage"
	EnvironmentDevelopment = "dev"
)

type Config struct {
	HTTP       HTTPConfig
	Auth       AuthConfig
	PG         DBConfig
	Redis      RedisConfig
	App        AppConfig
	Brevo      BrevoConfig
	Vault      VaultConfig
	Client     ClientConfig
	Duels      Duels `envPrefix:"DUELS_"`
	ClickHouse ClickHouseConfig
}

type VaultConfig struct {
	RoleID       string `env:"VAULT_ROLE_ID,required"`
	SecretID     string `env:"VAULT_SECRET_ID,required"`
	VaultAddress string `env:"VAULT_ADDRESS,required"`
}

type HTTPConfig struct {
	PublicDomain        string `env:"HTTP_PUBLIC_DOMAIN,required"`
	Address             string `env:"HTTP_ADDRESS,required"`
	Host                string `env:"HTTP_HOST,required"`
	Port                string `env:"HTTP_PORT,required"`
	SwaggerValidatorURL string `env:"SWAGGER_VALIDATOR_URL,required"`
	AllowOrigins        string `env:"ALLOW_ORIGINS,required"`
	AllowCredentials    bool   `env:"ALLOW_CREDENTIALS,required"`
}

type AuthConfig struct {
	SecretSignKey   string        `env:"SECRET_SIGN_KEY,required"`
	RefreshTokenTTL time.Duration `env:"REFRESH_TOKEN_TTL,required"`
	AccessTokenTTL  time.Duration `env:"ACCESS_TOKEN_TTL,required"`
	CodeTTL         time.Duration `env:"CODE_TTL" envDefault:"10m"`
	TurnstileSecret string        `env:"TURNSTILE_SECRET"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST,required"`
	DB       int    `env:"REDIS_DB,required"`
	Port     string `env:"REDIS_PORT,required"`
	Password string `env:"REDIS_PASSWORD"`
}

type BrevoConfig struct {
	BrevoSecretKey string `env:"BREVO_SECRET_KEY,required"`
	BrevoEmail     string `env:"BREVO_EMAIL,required"`
	BrevoName      string `env:"BREVO_NAME,required"`
}

type ClientConfig struct {
	JupiterBaseURL        string `env:"JUPITER_BASE_URL,required"`
	JupiterAPIKey         string `env:"JUPITER_API_KEY,required"`
	SolscanAPIKey         string `env:"SOLSCAN_API_KEY,required"`
	MoralisAPIKey         string `env:"MORALIS_API_KEY,required"`
	WalletAuthorityAPIKey string `env:"WALLET_AUTHORITY_API_KEY,required"`
}

type Duels struct {
	ResolveJoinNotBefore time.Duration `env:"RESOLVE_JOIN_NOT_BEFORE,required"`
}

type AppConfig struct {
	Environment string `env:"ENVIRONMENT,required"`
	CMCApiKey   string `env:"CMC_API_KEY,required"`

	SolanaURL                    string `env:"SOLANA_URL,required"`
	SolanaWSURL                  string `env:"SOLANA_WS_URL,required"`
	SolanaQuickNodeAPI           string `env:"SOLANA_QUICK_NODE_API"`
	SolanaPriorityUpdateInterval string `env:"SOLANA_PRIORITY_UPDATE_INTERVAL,required"`
	ContractAddress              string `env:"CONTRACT_ADDRESS,required"`
	ContractAddressAPI           string `env:"CONTRACT_ADDRESS_API,required"`

	FirebaseFilePath string `env:"FIREBASE_FILE_PATH,required"`

	WalletCacheEncryptionKey string        `env:"WALLET_CACHE_ENCRYPTION_KEY,required"`
	WalletCacheTTL           time.Duration `env:"WALLET_CACHE_TTL,required"`

	ShareImageAPI string `env:"SHARE_IMAGE_API,required"`

	MetricsUserRegistrationsSyncInterval string `env:"METRICS_USER_REGISTRATIONS_SYNC_INTERVAL" envDefault:"@every 15m"`
	DailyReportCron                      string `env:"DAILY_REPORT_CRON" envDefault:"@daily"`

	DuelQueueSize uint32 `env:"DUEL_QUEUE_SIZE,required"`

	DDProfitWalletAddress string `env:"DD_PROFIT_WALLET_ADDRESS,required"`
}

type ClickHouseConfig struct {
	Host     string `env:"CLICKHOUSE_HOST,required"`
	Port     int    `env:"CLICKHOUSE_PORT,required"`
	User     string `env:"CLICKHOUSE_USER,required"`
	Password string `env:"CLICKHOUSE_PASSWORD,required"`
	Database string `env:"CLICKHOUSE_DATABASE,required"`
}
