package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/pressly/goose/v3"
)

type Migrator struct {
	provider *goose.Provider
}

func New(provider *goose.Provider) *Migrator {
	return &Migrator{provider: provider}
}

func (m *Migrator) Up(ctx context.Context) error {
	mCtx, mCancel := context.WithTimeout(ctx, 1*time.Minute)
	defer mCancel()

	if _, err := m.provider.Up(mCtx); err != nil {
		return fmt.Errorf("migrator: provider up: %w", err)
	}
	return nil
}
