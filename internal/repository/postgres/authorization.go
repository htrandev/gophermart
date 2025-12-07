package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/htrandev/gophermart/internal/domain"
)

func (r *Repository) Register(ctx context.Context, user domain.User) (string, error) {
	var id uuid.UUID

	query := `INSERT INTO users (login, password)
		VALUES($1, $2)
	RETURNING id;`

	err := r.db.QueryRowContext(ctx, query,
		user.Login,
		user.HashPassword,
	).Scan(&id)
	if err != nil {
		if isUniqueErr(err) {
			return "", domain.ErrNotUniqueLogin
		}
		return "", fmt.Errorf("repository/register: scan returned id: %w", err)
	}

	return id.String(), nil
}

func (r *Repository) Login(ctx context.Context, login string) (domain.User, error) {
	var u domain.User

	query := `SELECT 
		id, login, password
	FROM users
	WHERE login = $1;`

	err := r.db.QueryRowContext(ctx, query, login).Scan(
		&u.Id,
		&u.Login,
		&u.HashPassword,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf("repository/login: scan: %w", err)
	}

	return u, nil
}
