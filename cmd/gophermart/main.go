package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"go.uber.org/zap"

	pgmigration "github.com/htrandev/gophermart/deployments/migrations/postgres"
	"github.com/htrandev/gophermart/internal/authorizer"
	"github.com/htrandev/gophermart/internal/clients/accrual"
	"github.com/htrandev/gophermart/internal/handler"
	"github.com/htrandev/gophermart/internal/repository/postgres"
	"github.com/htrandev/gophermart/internal/repository/postgres/migration"
	"github.com/htrandev/gophermart/internal/router"
	"github.com/htrandev/gophermart/internal/service"
	"github.com/htrandev/gophermart/pkg/logger"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("run ends with error: %s", err.Error())
	}
}

func run() error {
	flags, err := parseFlags()
	if err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}

	log.Println("init logger")
	zl, err := logger.NewZapLogger(flags.logLvl)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	zl.Info("init db")
	db, err := sql.Open("pgx", flags.databaseURI)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	pingCtx, pingCancel := context.WithTimeout(ctx, 5*time.Second)
	defer pingCancel()

	zl.Info("ping db")
	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping db: %w. dsn: [%s]", err, flags.databaseURI)
	}

	zl.Info("init provider")
	provider, err := goose.NewProvider(database.DialectPostgres, db, pgmigration.Embed)
	if err != nil {
		return fmt.Errorf("goose: create new provider: %w", err)
	}

	zl.Info("init migrator")
	migrator := migration.New(provider)

	zl.Info("run migrations")
	if err := migrator.Up(ctx); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	zl.Info("init repository")
	repository := postgres.NewRepository(db)

	zl.Info("init authorizer")
	authorizer := authorizer.New(flags.key, flags.tokenTTL)

	zl.Info("init resty client")
	rc := resty.New().
		SetTimeout(30 * time.Second)

	zl.Info("init client")
	client := accrual.NewClient(&accrual.ClientOptions{
		MaxRetry: 3,
		Addr:     flags.accrualSystemAddr,
		Client:   rc,
		Logger:   zl,
	})

	zl.Info("init service")
	service := service.New(&service.ServiceOptions{
		Authorizer: authorizer,
		Repository: repository,
		Client:     client,
		NumWorkers: 3,
		Logger:     zl,
	})
	service.Run(ctx)
	defer service.Close()

	zl.Info("init handler")
	handler := handler.New(&handler.HandlerOptions{
		Logger:  zl,
		Service: service,
	})

	zl.Info("init router")
	router, err := router.New(authorizer, handler, zl)
	if err != nil {
		return fmt.Errorf("can't create new router: %w", err)
	}

	srv := http.Server{
		Addr:    flags.addr,
		Handler: router,
	}

	zl.Info("start serving", zap.String("addr", flags.addr))
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("can't start server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	shutDownCtx, shutDownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutDownCancel()

	if err := srv.Shutdown(shutDownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}
