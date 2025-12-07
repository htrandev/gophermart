package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) Truncate(ctx context.Context) error {
	if _, err := r.db.ExecContext(ctx, `TRUNCATE TABLE orders`); err != nil {
		return fmt.Errorf("repository/truncate orders: exec: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `TRUNCATE TABLE users CASCADE;`); err != nil {
		return fmt.Errorf("repository/truncate users: exec: %w", err)
	}
	return nil
}

func isUniqueErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return true
	}
	return false
}
