package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/htrandev/gophermart/internal/domain"
)

func (r *Repository) CreateOrder(ctx context.Context, order domain.Order) error {
	selectQuery := `SELECT user_id FROM orders WHERE number = $1;`
	createQuery := `INSERT INTO orders(number, user_id) VALUES ($1, $2);`

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("repository/createOrder: begin transaction: %w", err)
	}
	defer tx.Rollback()

	var userID uuid.UUID
	if err := tx.QueryRowContext(ctx, selectQuery, order.Number).Scan(&userID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("repository/createOrder: select scan: %w", err)
		}
	}

	switch userID {
	// если заказа не было
	case uuid.Nil:
	// если пользователь уже загрузил этот заказ
	case order.UserID:
		return domain.ErrOrderAlreadyAddToUser
	// если заказ был загружен другим пользователем
	default:
		return domain.ErrOrderCreatedByAnotherUser
	}

	if _, err := tx.ExecContext(ctx, createQuery, order.Number, order.UserID); err != nil {
		return fmt.Errorf("repository/createOrder: create new order: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository/createOrder: commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) GetOrders(ctx context.Context, userID uuid.UUID) ([]domain.Order, error) {
	query := `SELECT 
		number, 
		status, 
		created_at, 
		accrual 
	FROM orders 
	WHERE user_id = $1
	ORDER BY created_at DESC
	;`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("repository/getOrders: query: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order

	for rows.Next() {
		var order domain.Order
		var status string

		if err := rows.Scan(
			&order.Number,
			&status,
			&order.CreatedAt,
			&order.Accrual,
		); err != nil {
			return nil, fmt.Errorf("repository/getOrders: scan: %w", err)
		}

		order.UserID = userID
		order.Status = domain.ConvertStringToOrderStatus(status)

		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("repository/getOrders: rows err: %w", err)
	}

	if len(orders) == 0 {
		return nil, domain.ErrNotFound
	}
	return orders, nil
}

func (r *Repository) UpdateOrder(ctx context.Context, order domain.Order) error {
	if order.Status == domain.OrderStatusProcessed {
		if err := r.processOrder(ctx, order); err != nil {
			return fmt.Errorf("process order: %w", err)
		}
		return nil
	}

	if err := r.updateStatus(ctx, order); err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return nil
}

func (r *Repository) processOrder(ctx context.Context, order domain.Order) error {
	orderQuery := `UPDATE orders 
		SET accrual = $1, status = $2, updated_at = now()
		WHERE number = $3
	;`

	userQuery := `UPDATE users 
		SET current_balance = current_balance + $1
		WHERE id = $2
	;`

	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return fmt.Errorf("repository/processOrder: begin transaction")
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, orderQuery,
		order.Accrual,
		order.Status,
		order.Number,
	); err != nil {
		return fmt.Errorf("update order [%s] info: %w", order.Number, err)
	}

	if _, err := tx.ExecContext(ctx, userQuery,
		order.Accrual,
		order.UserID,
	); err != nil {
		return fmt.Errorf("update user [%s] balance for order [%s]: %w", order.UserID, order.Number, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("repository/processOrder: commit transaction: %w", err)
	}
	return nil
}

func (r *Repository) updateStatus(ctx context.Context, order domain.Order) error {
	orderQuery := `UPDATE orders SET status = $1, updated_at = now() WHERE number = $2;`
	if _, err := r.db.ExecContext(ctx, orderQuery,
		order.Status,
		order.Number,
	); err != nil {
		return fmt.Errorf("update order [%s] status: %w", order.Number, err)
	}
	return nil
}
