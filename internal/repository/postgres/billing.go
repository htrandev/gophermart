package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
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
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("repository/withdraw: begin transaction: %w", err)
	}
	defer tx.Rollback()

	chechOrderQuery := `SELECT id FROM orders WHERE number = $1;`
	var orderID uuid.UUID
	if err := tx.QueryRowContext(ctx, chechOrderQuery, withdraw.Order).Scan(&orderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("repository/withdraw: check order existence: %w", err)
	}

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
