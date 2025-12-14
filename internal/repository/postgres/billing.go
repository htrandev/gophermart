package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/htrandev/gophermart/internal/domain"
)

func (r *Repository) GetBalance(ctx context.Context, userID uuid.UUID) (domain.Balance, error) {
	var balance domain.Balance

	query := `SELECT current_balance, withdrawn FROM users WHERE id = $1`
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&balance.Balance,
		&balance.Withdrawn,
	); err != nil {
		return domain.Balance{}, fmt.Errorf("repository/getBalance: scan: %w", err)
	}

	return balance, nil
}

func (r *Repository) Withdraw(ctx context.Context, withdraw domain.WithdrawRequest) error {
	err := r.withdraw(ctx, withdraw)
	if err != nil {
		if isRetryable(err) {
			for i := 0; i < r.maxRetry; i++ {
				delay := i*2 + 1
				time.Sleep(time.Second * time.Duration(delay))
				if err := r.withdraw(ctx, withdraw); err != nil {
					if isRetryable(err) {
						continue
					}
					return fmt.Errorf("repository/withdrawn: withdraw retry: %d: unretriable: %w", i+1, err)
				}
				return nil
			}
			return fmt.Errorf("repository/withdrawn: reach retry limits: %w", err)
		}
		return fmt.Errorf("repository/withdrawn: withdraw: unretriable: %w", err)
	}
	return nil
}

func (r *Repository) withdraw(ctx context.Context, withdraw domain.WithdrawRequest) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("repository/withdraw: begin transaction: %w", err)
	}
	defer tx.Rollback()

	getUserBalanceQuery := `SELECT current_balance FROM users WHERE id = $1;`
	var balance float64
	if err := tx.QueryRowContext(ctx, getUserBalanceQuery, withdraw.UserID).Scan(&balance); err != nil {
		return fmt.Errorf("repository/withdraw: get user balance: %w", err)
	}
	if balance < withdraw.Sum {
		return domain.ErrNotEnoughPoints
	}

	addWithdrawQuery := `INSERT INTO withdrawals (order_number, sum, user_id)
		VALUES ($1, $2, $3)
	;`
	if _, err := tx.ExecContext(ctx, addWithdrawQuery, withdraw.Order, withdraw.Sum, withdraw.UserID); err != nil {
		return fmt.Errorf("repository/withdraw: add withdraw: %w", err)
	}

	userWithdrawQuery := `UPDATE users 
		SET current_balance = current_balance - $1, withdrawn = withdrawn + $1
		WHERE id = $2
	;`
	if _, err := tx.ExecContext(ctx, userWithdrawQuery, withdraw.Sum, withdraw.UserID); err != nil {
		return fmt.Errorf("repository/withdraw: update user balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository/withdraw: commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) GetWithdrawals(ctx context.Context, userID uuid.UUID) ([]domain.Withdraw, error) {
	query := `SELECT order_number, sum, created_at
	FROM withdrawals WHERE user_id = $1
	ORDER BY created_at DESC
	;`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository/getWithdrawals: query: %w", err)
	}
	defer rows.Close()

	var withdrawals []domain.Withdraw

	for rows.Next() {
		var withdraw domain.Withdraw

		if err := rows.Scan(
			&withdraw.Order,
			&withdraw.Sum,
			&withdraw.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository/getWithdrawals: scan: %w", err)
		}
		withdrawals = append(withdrawals, withdraw)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("repository/getWithdrawals: rows err: %w", err)
	}

	if len(withdrawals) == 0 {
		return nil, domain.ErrNotFound
	}
	return withdrawals, nil
}

func isSerializationFailure(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.SerializationFailure
}

func isPgConnErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
		return true
	}
	return false
}

func isRetryable(err error) bool {
	return isPgConnErr(err) || isSerializationFailure(err)
}
