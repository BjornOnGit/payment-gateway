package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/BjornOnGit/payment-gateway/internal/model"
)

type SettlementRepository interface {
	CreateSettlementAttempt(ctx context.Context, s *model.Settlements) error
	GetByID(ctx context.Context, id string) (*model.Settlements, error)
	List(ctx context.Context, limit int, offset int) ([]*model.Settlements, error)
}

type PostgresSettlementRepository struct {
	db *sql.DB
}

func NewPostgresSettlementRepository(db *sql.DB) *PostgresSettlementRepository {
	return &PostgresSettlementRepository{db: db}
}

func (r *PostgresSettlementRepository) CreateSettlementAttempt(ctx context.Context, s *model.Settlements) error {
	query := `
		INSERT INTO settlements (id, merchant_account_id, external_reference, status, amount, metadata, attempts, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	// Marshal metadata to JSON
	metadataJSON, err := json.Marshal(s.Metadata)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		s.ID,
		s.MerchantAccountID,
		s.ExternalReference,
		s.Status,
		s.Amount,
		metadataJSON,
		s.Attempts,
		s.CreatedAt,
		s.UpdatedAt,
	)
	return err
}

func (r *PostgresSettlementRepository) GetByID(
	ctx context.Context,
	id string,
) (*model.Settlements, error) {
	query := `
		SELECT
			id, merchant_account_id, external_reference, status, amount,
			metadata, attempts, created_at, updated_at
		FROM settlements
		WHERE id = $1
	`

	var s model.Settlements
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID,
		&s.MerchantAccountID,
		&s.ExternalReference,
		&s.Status,
		&s.Amount,
		&metadataJSON,
		&s.Attempts,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal metadata from JSON
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &s.Metadata); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

func (r *PostgresSettlementRepository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*model.Settlements, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT
			id, merchant_account_id, external_reference, status, amount,
			metadata, attempts, created_at, updated_at
		FROM settlements
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settlements []*model.Settlements

	for rows.Next() {
		var s model.Settlements
		var metadataJSON []byte

		err := rows.Scan(
			&s.ID,
			&s.MerchantAccountID,
			&s.ExternalReference,
			&s.Status,
			&s.Amount,
			&metadataJSON,
			&s.Attempts,
			&s.CreatedAt,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal metadata from JSON
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &s.Metadata); err != nil {
				return nil, err
			}
		}

		settlements = append(settlements, &s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return settlements, nil
}
