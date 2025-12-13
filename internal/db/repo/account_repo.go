package repo

import (
	"context"
	"database/sql"

	"github.com/BjornOnGit/payment-gateway/internal/model"
	"github.com/google/uuid"
)

type AccountRepository interface {
	GetOrCreateMerchantAccount(ctx context.Context, merchantID uuid.UUID, currency string) (*model.Accounts, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID, accountType model.AccountType) (*model.Accounts, error)
}

type PostgresAccountRepository struct {
	db *sql.DB
}

func NewPostgresAccountRepository(db *sql.DB) *PostgresAccountRepository {
	return &PostgresAccountRepository{db: db}
}

// GetOrCreateMerchantAccount retrieves or creates a merchant account
func (r *PostgresAccountRepository) GetOrCreateMerchantAccount(
	ctx context.Context,
	merchantID uuid.UUID,
	currency string,
) (*model.Accounts, error) {
	// Try to get existing account
	account, err := r.GetByOwnerID(ctx, merchantID, model.AccountTypeMerchant)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// If account exists, return it
	if account != nil {
		return account, nil
	}

	// Create new merchant account
	newAccount := &model.Accounts{
		ID:          uuid.New(),
		OwnerID:     merchantID,
		AccountType: model.AccountTypeMerchant,
		Currency:    currency,
	}

	query := `
		INSERT INTO accounts (id, owner_id, owner_type, account_type, currency, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err = r.db.QueryRowContext(
		ctx,
		query,
		newAccount.ID,
		newAccount.OwnerID,
		"merchant",
		newAccount.AccountType,
		newAccount.Currency,
	).Scan(&newAccount.CreatedAt, &newAccount.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return newAccount, nil
}

// GetByOwnerID retrieves an account by owner ID and account type
func (r *PostgresAccountRepository) GetByOwnerID(
	ctx context.Context,
	ownerID uuid.UUID,
	accountType model.AccountType,
) (*model.Accounts, error) {
	query := `
		SELECT id, owner_id, account_type, currency, created_at, updated_at
		FROM accounts
		WHERE owner_id = $1 AND account_type = $2
		LIMIT 1
	`

	var account model.Accounts
	err := r.db.QueryRowContext(ctx, query, ownerID, accountType).Scan(
		&account.ID,
		&account.OwnerID,
		&account.AccountType,
		&account.Currency,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &account, nil
}
