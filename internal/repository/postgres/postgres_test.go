package postgres_test

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/htrandev/gophermart/internal/repository/postgres"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type RepositorySuite struct {
	suite.Suite

	db         *sql.DB
	repository *postgres.Repository
}

func (s *RepositorySuite) SetupSuite() {
	databaseURI := "postgresql://postgres:postgres@postgres/praktikum?sslmode=disable"
	if dbURI := os.Getenv("DATABASE_URI"); dbURI != "" {
		databaseURI = dbURI
	}

	db, err := sql.Open("pgx", databaseURI)
	s.Require().NoError(err)

	s.db = db
	s.repository = postgres.NewRepository(db, 3)
}

func (s *RepositorySuite) TearDownSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s.repository.Truncate(ctx)
	s.repository.Close()
}
