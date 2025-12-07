package postgres_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/htrandev/gophermart/internal/domain"
	"github.com/stretchr/testify/suite"
)

type BillingSuite struct {
	RepositorySuite
}

func (s *BillingSuite) SetupSuite() {
	s.RepositorySuite.SetupSuite()
}

// Чистим базу после каждого теста, чтобы не было ошибки уникальности логина.
func (s *BillingSuite) TearDownTest() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.RepositorySuite.repository.Truncate(ctx)
}

func TestBilling(t *testing.T) {
	suite.Run(t, new(BillingSuite))
}

func (s *BillingSuite) TearDownSuite() {
	s.RepositorySuite.TearDownSuite()
}

func (s *BillingSuite) TestGetBalance() {
	ctx := context.Background()

	user := domain.User{
		Login:        "test",
		HashPassword: "password",
	}

	id, err := s.repository.Register(ctx, user)
	s.Require().NoError(err)
	userID, err := uuid.Parse(id)
	s.Require().NoError(err)

	s.Run("valid", func() {
		b, err := s.repository.GetBalance(ctx, userID)
		s.Require().NoError(err)
		s.Require().Equal(domain.Balance{}, b)
	})

	s.Run("context cancel", func() {
		newCtx, cancel := context.WithCancel(ctx)
		cancel()

		_, err := s.repository.GetBalance(newCtx, userID)
		s.Require().Error(err)

	})
}

func (s *BillingSuite) TestWithdraw() {
	ctx := context.Background()

	orderID := uuid.New()
	order := domain.Order{
		ID:     orderID,
		Number: "test",
	}

	var userID uuid.UUID
	addUserQuery := `INSERT INTO users (login, password, current_balance) VALUES ('test', 'password', 1000) RETURNING id;`
	err := s.db.QueryRowContext(ctx, addUserQuery).Scan(&userID)
	s.Require().NoError(err)

	addOrderQuery := `INSERT INTO orders (id, number, user_id) VALUES ($1, $2, $3);`
	_, err = s.db.ExecContext(ctx, addOrderQuery, order.ID, order.Number, userID)
	s.Require().NoError(err)

	s.Run("valid", func() {
		err := s.repository.Withdraw(ctx, domain.WithdrawRequest{UserID: userID, Order: order.Number, Sum: 100})
		s.Require().NoError(err)
	})

	s.Run("unknown order", func() {
		err := s.repository.Withdraw(ctx, domain.WithdrawRequest{UserID: userID, Order: "unknown order", Sum: 100})
		s.Require().ErrorIs(err, domain.ErrNotFound)
	})

	s.Run("unknown user", func() {
		err := s.repository.Withdraw(ctx, domain.WithdrawRequest{UserID: uuid.New(), Order: order.Number, Sum: 100})
		s.Require().ErrorIs(err, sql.ErrNoRows)
	})

	s.Run("not enough points", func() {
		err := s.repository.Withdraw(ctx, domain.WithdrawRequest{UserID: userID, Order: order.Number, Sum: 100000})
		s.Require().ErrorIs(err, domain.ErrNotEnoughPoints)
	})
}

func (s *BillingSuite) TestGetWithdrawals() {
	ctx := context.Background()

	var userID uuid.UUID
	addUserQuery := `INSERT INTO users (login, password, current_balance) VALUES ('test', 'password', 1000) RETURNING id;`
	err := s.db.QueryRowContext(ctx, addUserQuery).Scan(&userID)
	s.Require().NoError(err)

	withdraws := []domain.WithdrawRequest{
		{UserID: userID, Order: "test_1", Sum: 1},
		{UserID: userID, Order: "test_2", Sum: 2},
		{UserID: userID, Order: "test_3", Sum: 3},
	}

	addOrderQuery := `INSERT INTO orders (number, user_id) VALUES ($1, $2);`
	stmt, err := s.db.PrepareContext(ctx, addOrderQuery)
	s.Require().NoError(err)

	for _, withdraw := range withdraws {
		stmt.ExecContext(ctx, withdraw.Order, userID)
		err := s.repository.Withdraw(ctx, withdraw)
		s.Require().NoError(err)
	}

	s.Run("valid", func() {
		w, err := s.repository.GetWithdrawals(ctx, userID)
		s.Require().NoError(err)
		s.Require().Len(w, 3)

		for i := range withdraws {
			// тк получаем данные о списании в порядке от нового к старому,
			// то надо сравнивать со слайсом наоборот
			j := len(withdraws) - 1 - i
			s.Require().Equal(withdraws[j].Order, w[i].Order)
			s.Require().Equal(withdraws[j].Sum, w[i].Sum)
		}
	})

	s.Run("no withdrawals", func() {
		_, err := s.repository.GetWithdrawals(ctx, uuid.New())
		s.Require().Error(err, domain.ErrNotFound)
	})
}
