package click

import (
	"context"
	"database/sql"
	"dd-prediction-api/config"
	"errors"
	"fmt"
	"net/url"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-migrate/migrate/v4"
	goclickhouse "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/uptrace/go-clickhouse/ch"
)

const clickhouseMigrationDir = "migrations/clickhouse"

func CreateClickHouseConnection(c *config.Config) (*ch.DB, error) {
	ctx := context.Background()

	if err := ensureClickHouseDatabase(c); err != nil {
		return nil, err
	}

	db := ch.Connect(ch.WithAddr(fmt.Sprintf("%s:%d", c.ClickHouse.Host, c.ClickHouse.Port)),
		ch.WithUser(c.ClickHouse.User),
		ch.WithPassword(c.ClickHouse.Password),
		ch.WithDatabase(c.ClickHouse.Database),
		ch.WithInsecure(true),
	)

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	userInfo := url.UserPassword(c.ClickHouse.User, c.ClickHouse.Password)
	dsn := fmt.Sprintf("clickhouse://%s@%s:%d/%s?secure=false",
		userInfo.String(),
		c.ClickHouse.Host,
		c.ClickHouse.Port,
		c.ClickHouse.Database,
	)
	sqlDB, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sql.DB for migrations: %w", err)
	}

	if err = makeClickhouseMigration(sqlDB, clickhouseMigrationDir, c.ClickHouse.Database); err != nil {
		sqlDB.Close()
		return nil, err
	}

	sqlDB.Close()

	return db, nil
}

func ensureClickHouseDatabase(c *config.Config) error {
	userInfo := url.UserPassword(c.ClickHouse.User, c.ClickHouse.Password)
	dsn := fmt.Sprintf(
		"clickhouse://%s@%s:%d/default?secure=false",
		userInfo.String(),
		c.ClickHouse.Host,
		c.ClickHouse.Port,
	)

	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		return fmt.Errorf("open default clickhouse db failed: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping default clickhouse db failed: %w", err)
	}

	if _, err := db.Exec(
		fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", c.ClickHouse.Database),
	); err != nil {
		return fmt.Errorf("create clickhouse database failed: %w", err)
	}

	return nil
}

func makeClickhouseMigration(conn *sql.DB, migrationDir string, dbName string) error {
	driver, err := goclickhouse.WithInstance(conn, &goclickhouse.Config{
		MultiStatementEnabled: true,
	})
	if err != nil {
		return fmt.Errorf("clickhouse.WithInstance: %s", err.Error())
	}
	mg, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationDir),
		dbName, driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithDatabaseInstance: %s", err.Error())
	}
	if err = mg.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	}
	return nil
}
