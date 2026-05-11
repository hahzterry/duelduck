package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"dd-prediction-api/internal/model"
	"dd-prediction-api/pkg/repository"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type DuelRepository struct {
	repository.Generic[model.Duel, uuid.UUID]
}

type duelStatusHistory struct {
	bun.BaseModel `bun:"table:duel_status_history,alias:dsh"`

	DuelID    uuid.UUID `bun:"duel_id,notnull"`
	Status    uint8     `bun:"status,notnull"`
	ChangedAt time.Time `bun:"changed_at,notnull"`
}

func NewDuelRepository(
	genericRepository repository.Generic[model.Duel, uuid.UUID],
) *DuelRepository {
	return &DuelRepository{
		Generic: genericRepository,
	}
}

func (r *DuelRepository) WithTx(tx bun.Tx) *DuelRepository {
	return &DuelRepository{Generic: r.Generic.WithTx(tx)}
}

func (r *DuelRepository) insertStatusHistory(
	ctx context.Context,
	duelID uuid.UUID,
	status uint8,
	changedAt time.Time,
) error {
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	}

	_, err := r.DB.NewInsert().
		Model(&duelStatusHistory{
			DuelID:    duelID,
			Status:    status,
			ChangedAt: changedAt,
		}).
		Exec(ctx)
	return err
}

func (r *DuelRepository) Create(ctx context.Context, duel *model.Duel) error {
	_, err := r.DB.NewInsert().Model(duel).Exec(ctx)
	if err != nil {
		return err
	}

	return r.insertStatusHistory(ctx, duel.ID, duel.Status, time.Now().UTC())
}

func (r *DuelRepository) CreateBulk(ctx context.Context, duels []model.Duel) error {
	_, err := r.DB.NewInsert().Model(&duels).Exec(ctx)
	if err != nil {
		return err
	}

	history := make([]duelStatusHistory, 0, len(duels))
	for i := range duels {
		changedAt := duels[i].UpdatedAt
		if changedAt.IsZero() {
			changedAt = time.Now().UTC()
		}

		history = append(history, duelStatusHistory{
			DuelID:    duels[i].ID,
			Status:    duels[i].Status,
			ChangedAt: changedAt,
		})
	}

	if len(history) == 0 {
		return nil
	}

	_, err = r.DB.NewInsert().Model(&history).Exec(ctx)
	return err
}

func (r *DuelRepository) CountPublicProfileDuelsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (*model.PublicProfileDuelCounts, error) {
	out := &model.PublicProfileDuelCounts{}

	err := r.DB.NewRaw(`
		SELECT
			COUNT(*) FILTER (
				WHERE d.owner_id = ?
				  AND d.status IN (?, ?)
				  AND p.final_status = ?
			) AS created,
			COUNT(*) FILTER (
				WHERE d.owner_id <> ?
				  AND d.status IN (?, ?)
				  AND p.final_status = ?
			) AS playing,
			COUNT(*) FILTER (
				WHERE d.status = ?
				  AND d.final_result IS NOT NULL
				  AND p.final_status = ?
				  AND d.final_result = p.answer
			) AS won,
			COUNT(*) FILTER (
				WHERE d.status = ?
				  AND d.final_result IS NOT NULL
				  AND p.final_status = ?
				  AND d.final_result <> p.answer
			) AS lost
		FROM players p
		JOIN duels d ON d.id = p.duel_id
		WHERE p.user_id = ?
	`,
		// created
		userID,
		model.DuelStatusActive, model.DuelStatusWaitingForResolve,
		model.PlayerStatusActive,
		// playing
		userID,
		model.DuelStatusActive, model.DuelStatusWaitingForResolve,
		model.PlayerStatusActive,
		// won
		model.DuelStatusResolved,
		model.PlayerStatusResolved,
		// lost
		model.DuelStatusResolved,
		model.PlayerStatusResolved,
		// WHERE
		userID,
	).Scan(ctx, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Use only within TX
func (r *DuelRepository) Update(ctx context.Context, duel *model.Duel) error {
	var prevStatus uint8

	// select for update has an impact only in case Update() was invoked within TX
	err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		Column("status").
		Where("id = ?", duel.ID).
		For("UPDATE").
		Scan(ctx, &prevStatus)
	if err != nil {
		return err
	}

	_, err = r.DB.NewUpdate().
		Model(duel).
		OmitZero().
		WherePK().
		Exec(ctx)
	if err != nil {
		return err
	}

	if prevStatus == duel.Status {
		return nil
	}

	return r.insertStatusHistory(ctx, duel.ID, duel.Status, time.Now().UTC())
}

func (r *DuelRepository) GetAllDuels(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id")
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) CountDuelsFiltered(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) (int, error) {
	opts := &repository.Options{Filters: options.Filters}
	q := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id")
	q = opts.ApplyFilters(q)
	count, err := q.Count(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *DuelRepository) GetAllDuelsWhereParticipate(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("inner join players p on p.duel_id = duels.id and p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id")
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetDuelShowByID(
	ctx context.Context,
	userID uuid.UUID,
	duelID uuid.UUID,
) (*model.DuelShow, error) {
	duel := new(model.DuelShow)

	q := r.DB.NewSelect().
		Model(duel).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("duels.id = ?", duelID)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duel, nil
}

func (r *DuelRepository) GetDuelShowBySlug(
	ctx context.Context,
	userID uuid.UUID,
	slug string,
) (*model.DuelShow, error) {
	duel := new(model.DuelShow)

	q := r.DB.NewSelect().
		Model(duel).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("duels.slug = ?", slug)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duel, nil
}

func (r *DuelRepository) GetHistoryByUserID(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer as your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("p.user_id = ? and duels.status in (?, ?, ?)",
			userID,
			model.DuelStatusDenied,
			model.DuelStatusResolved,
			model.DuelStatusRefunded,
		)
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetUserDuels(
	ctx context.Context,
	userID uuid.UUID,
	options *repository.Options,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) as joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer as your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("left join players p on p.duel_id = duels.id AND p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("duels.owner_id = ?", userID)
	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetActiveUnresolvedCreatedDuelsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("inner join players p on p.duel_id = duels.id and p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("duels.owner_id = ?", userID).
		Where("duels.status IN (?)", bun.In([]uint8{model.DuelStatusActive, model.DuelStatusWaitingForResolve})).
		Where("duels.deadline > NOW()").
		Where("duels.resolved_at IS NULL").
		Where("duels.final_result IS NULL").
		Where("p.final_status = ?", model.PlayerStatusActive).
		Order("duels.created_at DESC")

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetActiveUnresolvedParticipatingDuelsExcludingOwner(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("inner join players p on p.duel_id = duels.id and p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("duels.owner_id <> ?", userID).
		Where("duels.status IN (?)", bun.In([]uint8{model.DuelStatusActive, model.DuelStatusWaitingForResolve})).
		Where("duels.deadline > NOW()").
		Where("duels.resolved_at IS NULL").
		Where("duels.final_result IS NULL").
		Where("p.final_status = ?", model.PlayerStatusActive).
		Order("duels.created_at DESC")

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetResolvedParticipatedDuelsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]model.DuelShow, error) {
	duels := make([]model.DuelShow, 0)

	q := r.DB.NewSelect().
		Model(&duels).
		ColumnExpr("duels.*").
		ColumnExpr("mdo.multi_duel_id").
		ColumnExpr("md.slug AS multi_duel_slug").
		ColumnExpr("(p.user_id IS NOT NULL) AS joined").
		ColumnExpr("p.final_status as player_status").
		ColumnExpr("p.answer AS your_answer").
		ColumnExpr("p.win_amount as win_amount").
		ColumnExpr("COALESCE(yes_counts.yes_count, 0) AS yes_count").
		ColumnExpr("duels.players_count - COALESCE(yes_counts.yes_count, 0) as no_count").
		ColumnExpr("u.image_url AS owner_image_url").
		Join("left join users u ON u.id = duels.owner_id").
		Join("left join multi_duel_outcomes mdo ON mdo.duel_id = duels.id").
		Join("left join (select id, slug from multi_duels) md ON md.id = mdo.multi_duel_id").
		Join("inner join players p on p.duel_id = duels.id and p.user_id = ?", userID).
		Join("left join (select duel_id, COUNT(*) AS yes_count FROM players WHERE answer = 1 GROUP BY duel_id) AS yes_counts ON yes_counts.duel_id = duels.id").
		Where("duels.status = ?", model.DuelStatusResolved).
		Where("duels.final_result IS NOT NULL").
		Order("duels.resolved_at DESC")

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) JoinDuel(
	ctx context.Context,
	userID uuid.UUID,
	req *model.JoinDuelReq,
	duel *model.Duel,
) (*model.Player, error) {
	var prevStatus uint8
	err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		Column("status").
		Where("id = ?", req.DuelID).
		Scan(ctx, &prevStatus)
	if err != nil {
		return nil, err
	}

	player := &model.Player{
		ID:        uuid.New(),
		UserID:    userID,
		DuelID:    req.DuelID,
		Answer:    req.Answer,
		WalletID:  req.WalletID,
		PaidPrice: req.PaidPrice,
		CreatedAt: time.Now(),
	}
	_, err = r.DB.NewInsert().Model(player).Exec(ctx)
	if err != nil {
		return nil, err
	}

	_, err = r.DB.NewUpdate().
		Model((*model.Duel)(nil)).
		Set("players_count = players_count + 1").
		Set("status = ?", duel.Status).
		Where("id = ?", req.DuelID).
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	if prevStatus != duel.Status {
		if err = r.insertStatusHistory(ctx, req.DuelID, duel.Status, time.Now().UTC()); err != nil {
			return nil, err
		}
	}

	return player, nil
}

func (r *DuelRepository) RefreshMaterializedView(
	ctx context.Context,
	viewName string,
) error {
	_, err := r.DB.NewRaw("REFRESH MATERIALIZED VIEW ?", bun.Ident(viewName)).Exec(ctx)

	return err
}

func (r *DuelRepository) GetOldDuelsInReview(
	ctx context.Context,
	symbol string,
) ([]model.Duel, error) {
	duels := make([]model.Duel, 0)

	err := r.DB.NewSelect().
		Model(&duels).
		Where("duels.symbol = ?", symbol).
		Where("duels.status = ?", model.DuelStatusActive).
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				WhereOr("duels.created_at < ?", time.Now().AddDate(0, 0, -7)).
				WhereOr("duels.deadline < ?", time.Now())
		}).
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) FindOwnerIDByDuelID(
	ctx context.Context,
	duelID uuid.UUID,
) (uuid.UUID, error) {
	var ownerID uuid.UUID

	err := r.DB.NewSelect().
		Table("duels").
		Column("owner_id").
		Where("id = ?", duelID).
		Scan(ctx, &ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, nil
		}
		return uuid.Nil, err
	}

	return ownerID, nil
}

func (r *DuelRepository) FindActiveDuelsWithDeadlineBetween(
	ctx context.Context,
	from time.Time,
	to time.Time,
) ([]*model.Duel, error) {
	var duels []*model.Duel

	err := r.DB.NewSelect().
		Model(&duels).
		Where("deadline BETWEEN ? AND ?", from, to).
		Where("status = ?", model.DuelStatusActive).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return duels, nil
}

func (r *DuelRepository) GetActiveDuelSymbols(ctx context.Context) ([]string, error) {
	var symbols []string
	err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		Column("symbol").
		Where("status = ?", model.DuelStatusActive).
		Where("deadline > NOW()").
		Distinct().
		Scan(ctx, &symbols)
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func (r *DuelRepository) CountDistinctParticipantsInActiveDuels(ctx context.Context) (int64, error) {
	var count int64
	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT players.user_id)").
		Join("JOIN duels ON duels.id = players.duel_id").
		Where("duels.status IN (?)", bun.In([]uint8{model.DuelStatusActive, model.DuelStatusWaitingForResolve})).
		Where("duels.deadline > NOW()").
		Scan(ctx, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *DuelRepository) CountDistinctParticipantsInActiveDuelsByCategory(ctx context.Context, categoryID uuid.UUID) (int64, error) {
	var count int64
	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("COUNT(DISTINCT players.user_id)").
		Join("JOIN duels ON duels.id = players.duel_id").
		Where("duels.status IN (?)", bun.In([]uint8{model.DuelStatusActive, model.DuelStatusWaitingForResolve})).
		Where("duels.deadline > NOW()").
		Where("duels.category_id = ?", categoryID).
		Scan(ctx, &count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *DuelRepository) GetActiveDuelSymbolsByCategory(ctx context.Context, categoryID uuid.UUID) ([]string, error) {
	var symbols []string
	err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		Column("symbol").
		Where("status = ?", model.DuelStatusActive).
		Where("deadline > NOW()").
		Where("category_id = ?", categoryID).
		Distinct().
		Scan(ctx, &symbols)
	if err != nil {
		return nil, err
	}
	return symbols, nil
}

func (r *DuelRepository) GetActiveDuelsPageByCategory(
	ctx context.Context,
	categoryID uuid.UUID,
	pageSize, page int,
) ([]*model.Duel, error) {
	if pageSize <= 0 || page <= 0 {
		return nil, nil
	}
	duels := make([]*model.Duel, 0, pageSize)
	err := r.DB.NewSelect().
		Model(&duels).
		Column("id", "symbol", "duel_price", "players_count").
		Where("status = ?", model.DuelStatusActive).
		Where("deadline > NOW()").
		Where("category_id = ?", categoryID).
		Order("id ASC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return duels, nil
}

func (r *DuelRepository) CountActiveDuelsByCategory(ctx context.Context, categoryID uuid.UUID) (int, error) {
	return r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		Where("status = ?", model.DuelStatusActive).
		Where("deadline > NOW()").
		Where("category_id = ?", categoryID).
		Count(ctx)
}

func (r *DuelRepository) GetActiveDuelTokenMints(
	ctx context.Context,
	excludeMints []string,
) ([]string, error) {
	var tokenMints []string
	q := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("DISTINCT st.mint").
		Join("JOIN solana_tokens AS st ON st.symbol = duels.symbol").
		Where("duels.status IN (?)", bun.In([]uint8{model.DuelStatusActive, model.DuelStatusWaitingForResolve}))

	if len(excludeMints) > 0 {
		q = q.Where("st.mint NOT IN (?)", bun.In(excludeMints))
	}

	err := q.Scan(ctx, &tokenMints)
	if err != nil {
		return nil, err
	}

	return tokenMints, nil
}

func (r *DuelRepository) GetActiveDuelTokenSymbols(
	ctx context.Context,
	isTournamentScope bool,
) ([]string, error) {

	var symbols []string

	q := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("DISTINCT symbol").
		Where("status IN (?)", bun.In([]uint8{
			model.DuelStatusActive,
			model.DuelStatusWaitingForResolve,
		}))

	if isTournamentScope {
		q = q.Where("tournament_id IS NOT NULL")
	}

	err := q.Scan(ctx, &symbols)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return symbols, nil
}

func (r *DuelRepository) GetActiveDuelSolanaTokens(
	ctx context.Context,
	options *repository.Options,
) ([]model.SearchSolanaTokens, error) {
	tokens := make([]model.SearchSolanaTokens, 0)

	activeDuelsSubq := r.DB.NewSelect().
		TableExpr("duels").
		ColumnExpr("1").
		Where("duels.symbol = st.symbol").
		Where("duels.status IN (?)", bun.In([]uint8{
			model.DuelStatusActive,
			model.DuelStatusWaitingForResolve,
		})).
		Where("duels.deadline > NOW()")

	q := r.DB.NewSelect().
		Model(&tokens).
		Where("EXISTS (?)", activeDuelsSubq)

	q = options.Apply(q)

	if err := q.Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return tokens, nil
		}
		return nil, err
	}

	return tokens, nil
}

func (r *DuelRepository) GetDuelSymbolsLast24Hours(ctx context.Context) ([]string, error) {
	var symbols []string

	err := r.DB.NewRaw(`
		SELECT DISTINCT symbol
		FROM (
			SELECT d.symbol
			FROM duels d
			WHERE d.created_at >= (NOW() - INTERVAL '24 hours')

			UNION

			SELECT d.symbol
			FROM duel_status_history dsh
			JOIN duels d ON d.id = dsh.duel_id
			WHERE dsh.changed_at >= (NOW() - INTERVAL '24 hours')
		) s
		ORDER BY symbol ASC
	`).Scan(ctx, &symbols)
	if err != nil {
		return nil, err
	}

	return symbols, nil
}

type DuelSymbolNativeVolumeRow struct {
	Symbol       string  `bun:"symbol"`
	NativeVolume float64 `bun:"native_volume"`
}

func (r *DuelRepository) GetActiveDuelsNativeWagerVolumeLast24Hours(
	ctx context.Context,
	symbols []string,
) ([]DuelSymbolNativeVolumeRow, error) {
	rows := make([]DuelSymbolNativeVolumeRow, 0)
	if len(symbols) == 0 {
		return rows, nil
	}

	err := r.DB.NewSelect().
		TableExpr("players AS p").
		Join("INNER JOIN duels d ON d.id = p.duel_id").
		ColumnExpr("d.symbol AS symbol").
		ColumnExpr("COALESCE(SUM(d.duel_price), 0) AS native_volume").
		Where("p.created_at >= (NOW() - INTERVAL '24 hours')").
		Where("d.status = ?", model.DuelStatusActive).
		Where("d.deadline > NOW()").
		Where("d.symbol IN (?)", bun.In(symbols)).
		GroupExpr("d.symbol").
		OrderExpr("d.symbol ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type DuelSymbolNativeFeeRow struct {
	Symbol    string  `bun:"symbol"`
	NativeFee float64 `bun:"native_fee"`
}

// native_fee = platform share of commission from resolved duels in last 24h.
// platform share is (duel_price * players_count) * (commission/100) / 2 per symbol.
func (r *DuelRepository) GetActiveDuelsNativePlatformFeeLast24Hours(
	ctx context.Context,
) ([]DuelSymbolNativeFeeRow, error) {
	rows := make([]DuelSymbolNativeFeeRow, 0)

	err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("symbol").
		ColumnExpr("COALESCE(SUM((duel_price * players_count) * (commission / 100.0) / 2.0), 0) AS native_fee").
		Where("status = ?", model.DuelStatusResolved).
		Where("resolved_at IS NOT NULL").
		Where("resolved_at >= (NOW() - INTERVAL '24 hours')").
		GroupExpr("symbol").
		OrderExpr("symbol ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type CreatorCommissionNativeRow struct {
	OwnerID       uuid.UUID `bun:"owner_id"`
	Symbol        string    `bun:"symbol"`
	NativeEarning float64   `bun:"native_earning"`
}

// native_earning = (duel_price * players_count) * (commission/100) / 2
func (r *DuelRepository) GetCreatorsCommissionNativeByResolvedDuelsLast24Hours(
	ctx context.Context,
) ([]CreatorCommissionNativeRow, error) {
	rows := make([]CreatorCommissionNativeRow, 0)

	err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("owner_id").
		ColumnExpr("symbol").
		ColumnExpr("COALESCE(SUM((duel_price * players_count) * (commission / 100.0) / 2.0), 0) AS native_earning").
		Where("status = ?", model.DuelStatusResolved).
		Where("resolved_at IS NOT NULL").
		Where("resolved_at >= (NOW() - INTERVAL '24 hours')").
		GroupExpr("owner_id, symbol").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type UserDuelWinsNativeBySymbolRow struct {
	Symbol    string  `bun:"symbol"`
	NativeWin float64 `bun:"native_win"`
}

// native_win = SUM(players.win_amount) in native token units for user's winning votes
// across resolved duels, grouped by duel symbol.
func (r *DuelRepository) GetUserDuelWinsNativeBySymbol(
	ctx context.Context,
	userID uuid.UUID,
) ([]UserDuelWinsNativeBySymbolRow, error) {
	rows := make([]UserDuelWinsNativeBySymbolRow, 0)

	err := r.DB.NewSelect().
		Model((*model.Player)(nil)).
		ColumnExpr("d.symbol AS symbol").
		ColumnExpr("COALESCE(SUM(players.win_amount), 0) AS native_win").
		Join("INNER JOIN duels d ON d.id = players.duel_id").
		Where("players.user_id = ?", userID).
		Where("players.final_status = ?", model.PlayerStatusResolved).
		Where("players.is_winner = ?", true).
		Where("d.status = ?", model.DuelStatusResolved).
		GroupExpr("d.symbol").
		OrderExpr("d.symbol ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type UserCreatorFeeNativeBySymbolRow struct {
	Symbol    string  `bun:"symbol"`
	NativeFee float64 `bun:"native_fee"`
}

// native_fee = SUM(duels.commission) in native token units for user's created resolved duels,
// grouped by duel symbol.
func (r *DuelRepository) GetUserCreatorFeeNativeBySymbol(
	ctx context.Context,
	userID uuid.UUID,
) ([]UserCreatorFeeNativeBySymbolRow, error) {
	rows := make([]UserCreatorFeeNativeBySymbolRow, 0)

	err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("symbol AS symbol").
		ColumnExpr("COALESCE(SUM(commission), 0) AS native_fee").
		Where("owner_id = ?", userID).
		Where("status = ?", model.DuelStatusResolved).
		GroupExpr("symbol").
		OrderExpr("symbol ASC").
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

type ProjectDuelStats struct {
	DuelCount      int     `bun:"duel_count"`
	TotalVolumeUSD float64 `bun:"total_volume_usd"`
}

func (r *DuelRepository) GetProjectDuelStats(ctx context.Context, projectID uuid.UUID) (*ProjectDuelStats, error) {
	var stats ProjectDuelStats
	err := r.DB.NewRaw(`
		SELECT
			COUNT(*)                                            AS duel_count,
			COALESCE(SUM(usd_price * players_count), 0)        AS total_volume_usd
		FROM duels
		WHERE project_id = ?
		  AND status = ?`,
		projectID, model.DuelStatusResolved,
	).Scan(ctx, &stats)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (r *DuelRepository) CountDistinctCreatorsLast24Hours(ctx context.Context) (int64, error) {
	var cnt int64
	err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		ColumnExpr("COUNT(DISTINCT owner_id)").
		Where("created_at >= (NOW() - INTERVAL '24 hours')").
		Scan(ctx, &cnt)
	return cnt, err
}

func (r *DuelRepository) GetTopResolvedDuelLast24HoursByVolume(ctx context.Context) (*model.Duel, error) {
	var d model.Duel
	err := r.DB.NewSelect().
		Model(&d).
		Where("status = ?", model.DuelStatusResolved).
		Where("resolved_at IS NOT NULL").
		Where("resolved_at >= (NOW() - INTERVAL '24 hours')").
		OrderExpr("(duel_price * players_count) DESC").
		OrderExpr("resolved_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DuelRepository) GetTopResolvedDuelLast24HoursByParticipants(ctx context.Context) (*model.Duel, error) {
	var d model.Duel
	err := r.DB.NewSelect().
		Model(&d).
		Where("status = ?", model.DuelStatusResolved).
		Where("resolved_at IS NOT NULL").
		Where("resolved_at >= (NOW() - INTERVAL '24 hours')").
		OrderExpr("players_count DESC").
		OrderExpr("(duel_price * players_count) DESC").
		OrderExpr("resolved_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

func (r *DuelRepository) ClearCategoryNameFromDuels(ctx context.Context, categoryID uuid.UUID) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Duel)(nil)).
		Set("category_name = NULL").
		Where("category_id = ?", categoryID).
		Exec(ctx)
	return err
}

func (r *DuelRepository) GetOwnerID(ctx context.Context, duelID uuid.UUID) (uuid.UUID, error) {
	var ownerID uuid.UUID
	err := r.DB.NewSelect().
		TableExpr("duels").
		ColumnExpr("owner_id").
		Where("id = ?", duelID).
		Scan(ctx, &ownerID)
	return ownerID, err
}

func (r *DuelRepository) SetRoomTokenPDA(
	ctx context.Context,
	duelID uuid.UUID,
	roomTokenPda string,
) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Duel)(nil)).
		Set("room_token_pda = ?", roomTokenPda).
		Where("id = ?", duelID).
		Exec(ctx)

	return err
}

func (r *DuelRepository) SetRefundedPlayersCount(
	ctx context.Context,
	duelID uuid.UUID,
	refundedCount uint64,
) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Duel)(nil)).
		Set("refunded_players_count = ?", refundedCount).
		Where("id = ?", duelID).
		Exec(ctx)

	return err
}

func (r *DuelRepository) SetNormalizedQuestionAndTopic(
	ctx context.Context,
	duelID uuid.UUID,
	question,
	topic string,
) error {
	_, err := r.DB.NewUpdate().
		Model((*model.Duel)(nil)).
		Set("question = ?", question).
		Set("topic = ?", topic).
		Where("id = ?", duelID).
		Exec(ctx)

	return err
}

type dunePredictionsDay struct {
	Date        string `bun:"date"`
	NewApproved int64  `bun:"new_approved"`
	Resolved    int64  `bun:"resolved"`
}

// DuneGetPredictionsDaily returns a CSV with daily counts of approved and resolved duels
// since 2026-01-01, filling every calendar day (0 when no events).
func (r *DuelRepository) DuneGetPredictionsDaily(ctx context.Context) (string, error) {
	var rows []dunePredictionsDay

	err := r.DB.NewRaw(`
		WITH date_series AS (
			SELECT to_char(generate_series('2026-01-01'::date, CURRENT_DATE, '1 day'), 'YYYY-MM-DD') AS day
		),
		events AS (
			SELECT
				to_char(changed_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS day,
				COUNT(*) FILTER (WHERE status = ?) AS new_approved,
				COUNT(*) FILTER (WHERE status = ?) AS resolved
			FROM duel_status_history
			WHERE status IN (?, ?)
			  AND changed_at >= '2026-01-01'
			GROUP BY to_char(changed_at AT TIME ZONE 'UTC', 'YYYY-MM-DD')
		)
		SELECT
			ds.day AS date,
			COALESCE(e.new_approved, 0) AS new_approved,
			COALESCE(e.resolved, 0)     AS resolved
		FROM date_series ds
		LEFT JOIN events e ON e.day = ds.day
		ORDER BY ds.day
	`,
		model.DuelStatusActive,
		model.DuelStatusResolved,
		model.DuelStatusActive,
		model.DuelStatusResolved,
	).Scan(ctx, &rows)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(`"date","new_approved","resolved"`)
	sb.WriteByte('\n')
	for _, r := range rows {
		fmt.Fprintf(&sb, "%q,%d,%d\n", r.Date, r.NewApproved, r.Resolved)
	}

	return strings.TrimRight(sb.String(), "\n"), nil
}

type duneTVLDay struct {
	Date    string  `bun:"date"`
	Symbol  string  `bun:"symbol"`
	TVL     float64 `bun:"tvl"`
	TVLUsdt float64 `bun:"-"`
}

func (r *DuelRepository) DuneGetTVLDaily(ctx context.Context, priceMap map[string]float64) (string, error) {
	var rows []duneTVLDay

	err := r.DB.NewRaw(`
		WITH duel_periods AS (
			SELECT
				d.id,
				d.duel_price,
				d.symbol,
				(MIN(CASE WHEN h.status = ? THEN h.changed_at END) AT TIME ZONE 'UTC')::date AS started_at,
				(MIN(CASE WHEN h.status IN (?, ?, ?) THEN h.changed_at END) AT TIME ZONE 'UTC')::date AS ended_at
			FROM duels d
			JOIN duel_status_history h ON h.duel_id = d.id
			WHERE h.status IN (?, ?, ?, ?)
			GROUP BY d.id, d.duel_price, d.symbol
			HAVING MIN(CASE WHEN h.status = ? THEN h.changed_at END) IS NOT NULL
		),
		date_series AS (
			SELECT generate_series('2026-01-01'::date, CURRENT_DATE, '1 day') AS day
		),
		tvl_per_day AS (
			SELECT
				ds.day,
				dp.symbol,
				SUM(dp.duel_price) AS tvl
			FROM date_series ds
			JOIN duel_periods dp
				ON dp.started_at <= ds.day
				AND (dp.ended_at IS NULL OR dp.ended_at > ds.day)
			JOIN players p
				ON p.duel_id = dp.id
				AND (p.created_at AT TIME ZONE 'UTC')::date <= ds.day
			GROUP BY ds.day, dp.symbol
		)
		SELECT
			to_char(day, 'YYYY-MM-DD') AS date,
			symbol,
			tvl
		FROM tvl_per_day
		ORDER BY day, symbol
	`,
		model.DuelStatusActive,
		model.DuelStatusDenied, model.DuelStatusResolved, model.DuelStatusRefunded,
		model.DuelStatusActive,
		model.DuelStatusDenied, model.DuelStatusResolved, model.DuelStatusRefunded,
		model.DuelStatusActive,
	).Scan(ctx, &rows)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(`"date","symbol","tvl","tvl_usdt"`)
	sb.WriteByte('\n')
	for _, r := range rows {
		r.TVLUsdt = r.TVL * priceMap[r.Symbol]
		fmt.Fprintf(&sb, "%q,%q,%.6f,%.6f\n", r.Date, r.Symbol, r.TVL, r.TVLUsdt)
	}

	return strings.TrimRight(sb.String(), "\n"), nil
}

type duneRevenueDay struct {
	Date        string  `bun:"date"`
	Symbol      string  `bun:"symbol"`
	Revenue     float64 `bun:"revenue"`
	RevenueUsdt float64 `bun:"-"`
}

// revenue = duel_price * players_count * commission / 200  (platform takes half of total commission).
func (r *DuelRepository) DuneGetRevenueDaily(ctx context.Context, priceMap map[string]float64) (string, error) {
	var rows []duneRevenueDay

	err := r.DB.NewRaw(`
		SELECT
			to_char(h.changed_at AT TIME ZONE 'UTC', 'YYYY-MM-DD') AS date,
			d.symbol,
			SUM(d.duel_price * d.players_count * d.commission / 200.0) AS revenue
		FROM duel_status_history h
		JOIN duels d ON d.id = h.duel_id
		WHERE h.status = ?
		  AND h.changed_at >= '2026-01-01'
		GROUP BY to_char(h.changed_at AT TIME ZONE 'UTC', 'YYYY-MM-DD'), d.symbol
		ORDER BY date, d.symbol
	`, model.DuelStatusResolved).Scan(ctx, &rows)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(`"date","symbol","revenue","revenue_usdt"`)
	sb.WriteByte('\n')
	for _, r := range rows {
		r.RevenueUsdt = r.Revenue * priceMap[r.Symbol]
		fmt.Fprintf(&sb, "%q,%q,%.6f,%.6f\n", r.Date, r.Symbol, r.Revenue, r.RevenueUsdt)
	}

	return strings.TrimRight(sb.String(), "\n"), nil
}

type dunePlatformSnapshot struct {
	Symbol                  string  `bun:"symbol"`
	PriceUSD                float64 `bun:"-"`
	TVL                     float64 `bun:"tvl"`
	TVLUsdt                 float64 `bun:"-"`
	ActivePredCount         int64   `bun:"active_pred_count"`
	ActiveUsers             int64   `bun:"active_users"`
	AlltimeRevenue          float64 `bun:"alltime_revenue"`
	AlltimeRevenueUsdt      float64 `bun:"-"`
	SelfResolvedShare       float64 `bun:"self_resolved_share"`
	SelfResolvedShareGlobal float64 `bun:"self_resolved_share_global"`
}

func (r *DuelRepository) DuneGetPlatformSnapshot(ctx context.Context, priceMap map[string]float64) (string, error) {
	var rows []dunePlatformSnapshot

	err := r.DB.NewRaw(`
		WITH active_duels AS (
			SELECT d.id, d.symbol, d.duel_price, d.players_count, d.is_owner_resolving
			FROM duels d
			WHERE d.status = ?
		),
		tvl_per_symbol AS (
			SELECT symbol, SUM(duel_price * players_count) AS tvl
			FROM active_duels
			GROUP BY symbol
		),
		active_pred_count AS (
			SELECT symbol, COUNT(*) AS cnt
			FROM active_duels
			GROUP BY symbol
		),
		active_users AS (
			SELECT d.symbol, COUNT(DISTINCT p.user_id) AS cnt
			FROM players p
			JOIN active_duels d ON d.id = p.duel_id
			GROUP BY d.symbol
		),
		alltime_revenue AS (
			SELECT d.symbol,
				SUM(d.duel_price * d.players_count * d.commission / 200.0) AS revenue
			FROM duels d
			WHERE d.status = ?
			GROUP BY d.symbol
		),
		self_resolved AS (
			SELECT
				symbol,
				CASE
					WHEN COUNT(*) = 0 THEN 0.0
					ELSE COUNT(*) FILTER (WHERE is_owner_resolving = true)::float / COUNT(*)
				END AS share
			FROM active_duels
			GROUP BY symbol
		),
		all_symbols AS (
			SELECT DISTINCT symbol FROM (
				SELECT symbol FROM tvl_per_symbol
				UNION SELECT symbol FROM alltime_revenue
			) s
		)
		SELECT
			s.symbol,
			COALESCE(t.tvl, 0)    AS tvl,
			COALESCE(ap.cnt, 0)   AS active_pred_count,
			COALESCE(au.cnt, 0)   AS active_users,
			COALESCE(ar.revenue, 0) AS alltime_revenue,
			COALESCE(sr.share, 0) AS self_resolved_share,
			(
				SELECT CASE
					WHEN COUNT(*) = 0 THEN 0.0
					ELSE ROUND(
						(COUNT(*) FILTER (WHERE is_owner_resolving = true)::float
							/ COUNT(*) * 100)::numeric,
						2
					)
				END
				FROM active_duels
			) AS self_resolved_share_global
		FROM all_symbols s
		LEFT JOIN tvl_per_symbol    t  ON t.symbol  = s.symbol
		LEFT JOIN active_pred_count ap ON ap.symbol = s.symbol
		LEFT JOIN active_users      au ON au.symbol = s.symbol
		LEFT JOIN alltime_revenue   ar ON ar.symbol = s.symbol
		LEFT JOIN self_resolved     sr ON sr.symbol = s.symbol
		ORDER BY s.symbol
	`, model.DuelStatusActive, model.DuelStatusResolved).Scan(ctx, &rows)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(`"symbol","price_usd","tvl","tvl_usdt","active_predictions","active_users","alltime_revenue","alltime_revenue_usdt","self_resolved_share","self_resolved_share_global"`)
	sb.WriteByte('\n')
	for _, r := range rows {
		r.PriceUSD = priceMap[r.Symbol]
		r.TVLUsdt = r.TVL * r.PriceUSD
		r.AlltimeRevenueUsdt = r.AlltimeRevenue * r.PriceUSD
		fmt.Fprintf(&sb, "%q,%.6f,%.6f,%.6f,%d,%d,%.6f,%.6f,%.6f,%.2f\n",
			r.Symbol, r.PriceUSD, r.TVL, r.TVLUsdt, r.ActivePredCount, r.ActiveUsers, r.AlltimeRevenue, r.AlltimeRevenueUsdt, r.SelfResolvedShare, r.SelfResolvedShareGlobal)
	}

	return strings.TrimRight(sb.String(), "\n"), nil
}

func (r *DuelRepository) GetUnresolvedDuelsByUserID(
	ctx context.Context,
	userID uuid.UUID,
) (int, error) {
	count, err := r.DB.NewSelect().
		Model((*model.Duel)(nil)).
		Join("JOIN players AS p ON p.duel_id = duels.id").
		Where("p.user_id = ?", userID).
		Where("duels.status = ?", model.DuelStatusWaitingForResolve).
		Where("duels.is_self_resolved = true").
		Count(ctx)
	if err != nil {
		return 0, err
	}
	return count, nil
}
