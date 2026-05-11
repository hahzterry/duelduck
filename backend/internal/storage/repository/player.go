package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/repository"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PlayerRepository struct {
	repository.Generic[model.Player, uuid.UUID]
}

func (r *PlayerRepository) CountDuelsPlayedByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var cnt int64
	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT duel_id)").
		Where("user_id = ?", userID).
		Scan(ctx, &cnt)
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r *PlayerRepository) CountUserParticipationsWithKnownResult(ctx context.Context, userID uuid.UUID) (int64, error) {
	var cnt int64
	err := r.DB.NewSelect().
		TableExpr("players AS p").
		ColumnExpr("COUNT(DISTINCT p.duel_id)").
		Join("INNER JOIN duels d ON d.id = p.duel_id").
		Where("p.user_id = ?", userID).
		Where("d.status = ?", model.DuelStatusResolved).
		Where("d.final_result IS NOT NULL").
		Scan(ctx, &cnt)
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

func (r *PlayerRepository) CountUserDuelWins(ctx context.Context, userID uuid.UUID) (int64, error) {
	var cnt int64
	err := r.DB.NewSelect().
		TableExpr("players AS p").
		ColumnExpr("COUNT(DISTINCT p.duel_id)").
		Join("INNER JOIN duels d ON d.id = p.duel_id").
		Where("p.user_id = ?", userID).
		Where("d.status = ?", model.DuelStatusResolved).
		Where("d.final_result IS NOT NULL").
		Where("p.is_winner = ?", true).
		Scan(ctx, &cnt)
	if err != nil {
		return 0, err
	}
	return cnt, nil
}

func NewPlayerRepository(
	genericRepository repository.Generic[model.Player, uuid.UUID],
) *PlayerRepository {
	return &PlayerRepository{
		Generic: genericRepository,
	}
}

func (r *PlayerRepository) CountDistinctPlayersLast24Hours(ctx context.Context) (int64, error) {
	var cnt int64
	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT user_id)").
		Where("created_at >= (NOW() - INTERVAL '24 hours')").
		Scan(ctx, &cnt)
	return cnt, err
}

func (r *PlayerRepository) WithTx(tx bun.Tx) *PlayerRepository {
	return &PlayerRepository{Generic: r.Generic.WithTx(tx)}
}

func (r *PlayerRepository) UserAlreadyParticipant(
	ctx context.Context,
	userID uuid.UUID,
	duelID uuid.UUID,
) (bool, error) {
	player := new(model.Player)

	ok, err := r.DB.NewSelect().
		Model(player).
		Where("user_id = ? and duel_id = ?", userID, duelID).
		Exists(ctx)
	if err != nil {
		return false, err
	}

	return ok, nil
}

func (r *PlayerRepository) GetAllPlayersByDuelID(
	ctx context.Context,
	duelID uuid.UUID,
	options *repository.Options,
) ([]model.PlayerShow, error) {
	players := make([]model.PlayerShow, 0)

	q := r.DB.NewSelect().
		Model(&players).
		ColumnExpr("distinct players.id, players.user_id, players.duel_id, players.answer, players.win_amount, players.final_status, players.is_winner, players.created_at, u.username, u.image_url").
		Join("inner join users u on u.id = players.user_id").
		Where("players.duel_id = ?", duelID)
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) GetDuelPlayersWithoutNotification(
	ctx context.Context,
	duelID uuid.UUID,
	notificationType uint8,
) ([]uuid.UUID, error) {
	var playerIDs []uuid.UUID

	err := r.DB.NewSelect().
		TableExpr("players as p").
		Column("p.user_id").
		Where("p.duel_id = ?", duelID).
		Where(`
        NOT EXISTS (
            SELECT 1
            FROM notifications n
            WHERE n.user_id = p.user_id
              AND n.notification_type = ?
              AND (n.data ->> 'duel_id')::uuid = ?
        )
    `, notificationType, duelID).
		Scan(ctx, &playerIDs)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return playerIDs, nil
}

func (r *PlayerRepository) GetDuelWinners(
	ctx context.Context,
	duelID uuid.UUID,
	correctAnswer uint8,
	deadline time.Time,
) ([]model.Player, error) {
	players := make([]model.Player, 0)

	err := r.DB.NewSelect().
		Model(&players).
		Where("duel_id = ?", duelID).
		Where("answer = ?", correctAnswer).
		Where("created_at <= ?", deadline).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) GetDuelLosersIDs(
	ctx context.Context,
	duelID uuid.UUID,
	wrongAnswer uint8,
	deadline time.Time,
) (uuid.UUIDs, error) {
	loserIDs := make(uuid.UUIDs, 0)

	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		Column("user_id").
		Where("duel_id = ?", duelID).
		Where("answer = ?", wrongAnswer).
		Where("created_at <= ?", deadline).
		Scan(ctx, &loserIDs)
	if err != nil {
		return nil, err
	}

	return loserIDs, nil
}

func (r *PlayerRepository) CountDuelWinners(
	ctx context.Context,
	duelID uuid.UUID,
	correctAnswer uint8,
	deadline time.Time,
) (int, error) {
	return r.DB.NewSelect().
		Model((*model.Player)(nil)).
		Where("duel_id = ?", duelID).
		Where("answer = ?", correctAnswer).
		Where("created_at <= ?", deadline).
		Count(ctx)
}

func (r *PlayerRepository) UpdateDuelWinners(
	ctx context.Context,
	winners []model.Player,
	winAmount float64,
) error {
	ids := make(uuid.UUIDs, 0, len(winners))
	for _, winner := range winners {
		ids = append(ids, winner.ID)
	}

	_, err := r.DB.NewUpdate().
		Model((*model.Player)(nil)).
		Where("players.id IN (?)", bun.In(ids)).
		Set("is_winner = ?", true).
		Set("win_amount = ?", winAmount).
		Set("final_status = ?", model.PlayerStatusResolved).
		Exec(ctx)

	return err
}

func (r *PlayerRepository) SetStatusToAll(
	ctx context.Context,
	duelID uuid.UUID,
	status uint8,
) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Player)(nil)).
		Where("players.duel_id = ?", duelID).
		Set("final_status = ?", status).
		Exec(ctx)

	return err
}

func (r *PlayerRepository) SetStatusToActiveByDuelID(
	ctx context.Context,
	duelID uuid.UUID,
	status uint8,
) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Player)(nil)).
		Where("players.duel_id = ?", duelID).
		Where("players.final_status = ?", model.PlayerStatusActive).
		Set("final_status = ?", status).
		Exec(ctx)

	return err
}

func (r *PlayerRepository) SetStatus(
	ctx context.Context,
	players []model.CryptoDuelPlayer,
	status uint8,
) error {
	data := r.DB.NewValues(&players)

	_, err := r.DB.NewUpdate().
		With("_data", data).
		Model((*model.Player)(nil)).
		TableExpr("_data").
		Where("players.id = _data.id").
		Set("final_status = ?", status).
		Exec(ctx)

	return err
}

func (r *PlayerRepository) MarkPlayersAsRefunded(
	ctx context.Context,
	duelID uuid.UUID,
	joinNotBefore time.Time,
) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Player)(nil)).
		Where("duel_id = ?", duelID).
		Where("created_at > ?", joinNotBefore).
		Set("final_status = ?", model.PlayerStatusRefunded).
		Exec(ctx)

	return err
}

func (r *PlayerRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
	duelID uuid.UUID,
) (*model.Player, error) {
	player := new(model.Player)

	err := r.DB.NewSelect().
		Model(player).
		Where("user_id = ?", userID).
		Where("duel_id = ?", duelID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return player, nil
}

func (r *PlayerRepository) GetCryptoDuelWinners(
	ctx context.Context,
	duelID uuid.UUID,
	answer uint8,
) ([]model.CryptoDuelPlayer, error) {
	players := make([]model.CryptoDuelPlayer, 0)

	err := r.DB.NewSelect().
		Model(&players).
		Column("players.id").
		Column("players.user_id").
		Column("players.paid_price").
		ColumnExpr("w.address AS public_address").
		Where("players.duel_id = ?", duelID).
		Where("players.answer = ?", answer).
		Where("players.final_status = ?", model.PlayerStatusActive).
		Join(`LEFT JOIN wallets AS w ON w.id = players.wallet_id`).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) GetCryptoDuelPlayers(
	ctx context.Context,
	duelID uuid.UUID,
) ([]model.CryptoDuelPlayer, error) {
	players := make([]model.CryptoDuelPlayer, 0)

	err := r.DB.NewSelect().
		Model(&players).
		Column("players.id").
		Column("players.user_id").
		Column("players.paid_price").
		ColumnExpr("w.address AS public_address").
		Where("players.duel_id = ?", duelID).
		Join(`LEFT JOIN wallets AS w ON w.id = players.wallet_id`).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) GetDuelPlayersToRefund(
	ctx context.Context,
	duelID uuid.UUID,
	votedAfter time.Time,
) ([]model.CryptoDuelPlayer, error) {
	players := make([]model.CryptoDuelPlayer, 0)

	err := r.DB.NewSelect().
		Model(&players).
		Column("players.id").
		Column("players.user_id").
		Column("players.answer").
		Column("players.paid_price").
		ColumnExpr("w.address AS public_address").
		Where("players.duel_id = ?", duelID).
		Where("players.final_status = ?", model.PlayerStatusActive).
		Where("players.created_at > ?", votedAfter).
		Join(`LEFT JOIN wallets AS w ON w.id = players.wallet_id`).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) FindDuelPlayersUserIDsToRefund(
	ctx context.Context,
	duelID uuid.UUID,
	deadline time.Time,
) ([]uuid.UUID, error) {
	var userIDs []uuid.UUID

	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		Column("user_id").
		Where("duel_id = ?", duelID).
		Where("final_status = ?", model.PlayerStatusActive).
		Where("created_at > ?", deadline).
		Scan(ctx, &userIDs)
	if err != nil {
		if userIDs == nil {
			return nil, nil
		}
		return nil, err
	}

	return userIDs, nil
}

func (r *PlayerRepository) HasUserJoinedMoreThanOneDuel(ctx context.Context, userID uuid.UUID) (bool, error) {
	var duelCount int

	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT duel_id)").
		Where("user_id = ?", userID).
		Scan(ctx, &duelCount)
	if err != nil {
		return false, err
	}

	return duelCount > 1, nil
}

func (r *PlayerRepository) GetRefundedPlayersByID(
	ctx context.Context,
	duelID uuid.UUID,
) ([]*model.Player, error) {
	var players []*model.Player

	err := r.DB.NewSelect().
		Model(&players).
		Where("duel_id = ?", duelID).
		Where("final_status = ?", model.PlayerStatusRefunded).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return players, nil
}

func (r *PlayerRepository) CountActiveUsers(
	ctx context.Context,
) (int, error) {
	var cnt int

	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT user_id)").
		Scan(ctx, &cnt)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}

	return cnt, err
}

type YesCountRow struct {
	DuelID   uuid.UUID `bun:"duel_id"`
	YesCount uint64    `bun:"yes_count"`
}

func (r *PlayerRepository) GetYesCountByDuelIDs(ctx context.Context, duelIDs []uuid.UUID) (map[uuid.UUID]uint64, error) {
	if len(duelIDs) == 0 {
		return nil, nil
	}
	var rows []YesCountRow
	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		Column("duel_id").
		ColumnExpr("COUNT(*) FILTER (WHERE answer = 1) AS yes_count").
		Where("duel_id IN (?)", bun.In(duelIDs)).
		Group("duel_id").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]uint64, len(rows))
	for _, row := range rows {
		out[row.DuelID] = row.YesCount
	}
	return out, nil
}

func (r *PlayerRepository) CountDuelPlayers(
	ctx context.Context,
	duelID uuid.UUID,
) (uint64, error) {
	count, err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		Where("duel_id = ?", duelID).
		Count(ctx)
	if err != nil {
		return 0, err
	}

	return uint64(count), nil
}

func (r *PlayerRepository) SumActivePaidPrice(ctx context.Context, duelID uuid.UUID, fallbackPrice float64) (float64, error) {
	var total float64
	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COALESCE(SUM(COALESCE(paid_price, ?)), 0)", fallbackPrice).
		Where("duel_id = ?", duelID).
		Where("final_status = ?", model.PlayerStatusActive).
		Scan(ctx, &total)
	return total, err
}

func (r *PlayerRepository) UpdateDuelWinnersVariableAmounts(ctx context.Context, winners []model.Player) error {
	_, err := r.DB.NewUpdate().
		Model(&winners).
		Column("is_winner", "win_amount", "final_status").
		Bulk().
		Exec(ctx)
	return err
}
