package repo

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/BjornOnGit/payment-gateway/internal/model"
	"github.com/google/uuid"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, tx *model.Transaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	List(ctx context.Context, limit int, offset int) ([]*model.Transaction, error)
}

type PostgresTransactionRepository struct {
	db *sql.DB
}

func NewPostgresTransactionRepository(db *sql.DB) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{db: db}
}

func (r *PostgresTransactionRepository) CreateTransaction(ctx context.Context, tx *model.Transaction) error {
	query := `
        INSERT INTO transactions (id, amount, currency, user_id, merchant_id, status, metadata, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `

	// Marshal metadata to JSON for database storage
	var metadataJSON []byte
	if tx.Metadata != nil {
		var err error
		metadataJSON, err = json.Marshal(tx.Metadata)
		if err != nil {
			return err
		}
	} else {
		metadataJSON = []byte("{}")
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		tx.ID,
		tx.Amount,
		tx.Currency,
		tx.UserID,
		tx.MerchantID,
		tx.Status,
		metadataJSON,
		tx.CreatedAt,
		tx.UpdatedAt,
	)
	return err
}

func (r *PostgresTransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `
        UPDATE transactions SET status = $1, updated_at = NOW() WHERE id = $2
    `

	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *PostgresTransactionRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*model.Transaction, error) {

	query := `
        SELECT 
            id, amount, currency, user_id, merchant_id, status,
            metadata, created_at, updated_at
        FROM transactions
        WHERE id = $1
    `

	var t model.Transaction
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID,
		&t.Amount,
		&t.Currency,
		&t.UserID,
		&t.MerchantID,
		&t.Status,
		&metadataJSON,
		&t.CreatedAt,
		&t.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Unmarshal metadata from JSON
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &t.Metadata); err != nil {
			return nil, err
		}
	}

	return &t, nil
}

func (r *PostgresTransactionRepository) List(
	ctx context.Context,
	limit int,
	offset int,
) ([]*model.Transaction, error) {
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
            id, amount, currency, user_id, merchant_id, status,
            metadata, created_at, updated_at
        FROM transactions
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*model.Transaction

	for rows.Next() {
		var t model.Transaction
		var metadataJSON []byte

		err := rows.Scan(
			&t.ID,
			&t.Amount,
			&t.Currency,
			&t.UserID,
			&t.MerchantID,
			&t.Status,
			&metadataJSON,
			&t.CreatedAt,
			&t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal metadata from JSON
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &t.Metadata); err != nil {
				return nil, err
			}
		}

		transactions = append(transactions, &t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
