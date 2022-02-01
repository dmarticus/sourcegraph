package connections

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/storetypes"
)

// memoryStore implements runner.Store but writes to migration metadata are
// not passed to any underlying persistence layer.
type memoryStore struct {
	db              *sql.DB
	appliedVersions []int
	pendingVersions []int
	failedVersions  []int
}

func newMemoryStore(db *sql.DB) runner.Store {
	return &memoryStore{
		db: db,
	}
}

func (s *memoryStore) Transact(ctx context.Context) (runner.Store, error) {
	return s, nil
}

func (s *memoryStore) Done(err error) error {
	return err
}

func (s *memoryStore) Versions(ctx context.Context) (appliedVersions, pendingVersions, failedVersions []int, _ error) {
	return s.appliedVersions, s.pendingVersions, s.failedVersions, nil
}

func (s *memoryStore) Lock(ctx context.Context) (bool, func(err error) error, error) {
	return true, func(err error) error { return err }, nil
}

func (s *memoryStore) TryLock(ctx context.Context) (bool, func(err error) error, error) {
	return true, func(err error) error { return err }, nil
}

func (s *memoryStore) Up(ctx context.Context, migration definition.Definition) error {
	return s.exec(ctx, migration, migration.UpQuery)
}

func (s *memoryStore) Down(ctx context.Context, migration definition.Definition) error {
	return s.exec(ctx, migration, migration.DownQuery)
}

func (s *memoryStore) WithMigrationLog(_ context.Context, _ definition.Definition, _ bool, f func() error) error {
	return f()
}

func (s *memoryStore) IndexStatus(_ context.Context, _, _ string) (storetypes.IndexStatus, bool, error) {
	return storetypes.IndexStatus{}, false, nil
}

func (s *memoryStore) exec(ctx context.Context, migration definition.Definition, query *sqlf.Query) error {
	_, err := s.db.ExecContext(ctx, query.Query(sqlf.PostgresBindVar), query.Args()...)
	if err != nil {
		s.failedVersions = append(s.failedVersions, migration.ID)
		return err
	}

	s.appliedVersions = append(s.appliedVersions, migration.ID)
	return nil
}
