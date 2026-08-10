package postgresadapter_test

import (
	"context"
	"errors"
	"flag"
	"go-platform-template/internal/adapter/postgresadapter"
	"go-platform-template/internal/core/apperror"
	"go-platform-template/internal/core/widget"
	"go-platform-template/internal/platform/config"
	"go-platform-template/internal/platform/pg"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var postgresContainer *tcpostgres.PostgresContainer

var pgPool *pgxpool.Pool

func restoreDB(t *testing.T) {
	t.Helper()
	ctx := t.Context()

	pgPool.Reset() // free connections so template copy can run
	if err := postgresContainer.Restore(ctx); err != nil {
		t.Fatalf("restore failed: %v", err)
	}
}

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(0)
	}
	os.Exit(run(m))
}

func run(m *testing.M) int {
	ctx := context.Background()

	var err error
	var (
		dbUser     = "TEST_POSTGRES_USER"
		dbPassword = "TEST_POSTGRES_PASS" // #nosec G101
		dbName     = "TEST_POSTGRES_DATABASE"
	)

	postgresContainer, err = tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase(dbName),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		tcpostgres.BasicWaitStrategies(),
		tcpostgres.WithSQLDriver("pgx"),
	)
	defer func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()
	if err != nil {
		log.Printf("failed to start container: %s", err)
		return 1
	}

	dbHost, err := postgresContainer.Host(ctx)
	if err != nil {
		return 1
	}
	port, err := postgresContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return 1
	}
	dbConfig := config.DatabaseConfig{
		User:     dbUser,
		Password: dbPassword,
		Host:     dbHost,
		Port:     int(port.Num()),
		Name:     dbName,
		SSLMode:  config.SSLDisable,
	}

	migration, err := migrate.New("file://../../../migrations", dbConfig.DSN())
	if err != nil {
		log.Printf("unexpected error on migration: %v", err)
		return 1
	}
	if err := migration.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Printf("unexpected error on migration: %v", err)
		return 1
	}

	srcErr, dbErr := migration.Close()
	if srcErr != nil || dbErr != nil {
		log.Printf("migration close: %v, %v", srcErr, dbErr)
		return 1
	}

	err = postgresContainer.Snapshot(ctx)
	if err != nil {
		log.Printf("unexpected error on taking snapshot of DB: %v", err)
		return 1
	}

	pgPool, err = pg.New(ctx, dbConfig, slog.Default())
	if err != nil {
		log.Printf("cannot create pool: %v", err)
		return 1
	}
	defer pgPool.Close()

	return m.Run()
}

func TestCreate(t *testing.T) {
	restoreDB(t)

	widgetRepo := postgresadapter.NewWidgetRepository(pgPool)

	wg := widget.Widget{
		Name:   "widget1",
		Weight: 1,
	}

	createdWidget, err := widgetRepo.Create(t.Context(), wg)
	if err != nil {
		t.Fatalf("unexpected error while creating widget: %v", err)
	}

	if createdWidget.ID <= 0 {
		t.Errorf("expected widget.id bigger than zero, got %d", createdWidget.ID)
	}

	if createdWidget.Name != wg.Name {
		t.Errorf("expected widget.name %q, got %q", wg.Name, createdWidget.Name)
	}

	if createdWidget.Weight != wg.Weight {
		t.Errorf("expected widget.weight %d, got %d", wg.Weight, createdWidget.Weight)
	}

	if createdWidget.CreatedAt.IsZero() {
		t.Errorf("expected widget.createdAt not zero")
	}

	if createdWidget.UpdatedAt.IsZero() {
		t.Errorf("expected widget.updatedAt not zero")
	}
}

func TestGet(t *testing.T) {
	restoreDB(t)

	widgetRepo := postgresadapter.NewWidgetRepository(pgPool)

	wg := widget.Widget{
		Name:   "widget1",
		Weight: 1,
	}

	createdWidget, err := widgetRepo.Create(t.Context(), wg)
	if err != nil {
		t.Fatalf("unexpected error while creating widget: %v", err)
	}

	fetchedWidget, err := widgetRepo.Get(t.Context(), createdWidget.ID)
	if err != nil {
		t.Fatalf("unexpected error while fetching widget: %v", err)
	}

	if fetchedWidget.ID != createdWidget.ID {
		t.Errorf("expected widget.id: %d, got %d", createdWidget.ID, fetchedWidget.ID)
	}

	if fetchedWidget.Name != createdWidget.Name {
		t.Errorf("expected widget.name %q, got %q", createdWidget.Name, fetchedWidget.Name)
	}

	if fetchedWidget.Weight != createdWidget.Weight {
		t.Errorf("expected widget.weight %d, got %d", createdWidget.Weight, fetchedWidget.Weight)
	}

	if !fetchedWidget.CreatedAt.Equal(createdWidget.CreatedAt) {
		t.Errorf("expected widget.createdat %q, got %q", createdWidget.CreatedAt, fetchedWidget.CreatedAt)
	}

	if !fetchedWidget.UpdatedAt.Equal(createdWidget.UpdatedAt) {
		t.Errorf("expected widget.updatedat %q, got %q", createdWidget.UpdatedAt, fetchedWidget.UpdatedAt)
	}
}

func TestUpdate(t *testing.T) {
	restoreDB(t)

	widgetRepo := postgresadapter.NewWidgetRepository(pgPool)

	wg := widget.Widget{
		Name:   "widget1",
		Weight: 1,
	}

	createdWidget, err := widgetRepo.Create(t.Context(), wg)
	if err != nil {
		t.Fatalf("unexpected error while creating widget: %v", err)
	}

	updatedWg := widget.Widget{
		ID:     createdWidget.ID,
		Name:   "updated_widget",
		Weight: 100,
	}

	updatedWidget, err := widgetRepo.Update(t.Context(), updatedWg)
	if err != nil {
		t.Fatalf("unexpected error while updating widget: %v", err)
	}

	if updatedWidget.ID != updatedWg.ID {
		t.Errorf("expected widget.id: %d, got %d", updatedWg.ID, updatedWidget.ID)
	}

	if updatedWidget.Name != updatedWg.Name {
		t.Errorf("expected widget.name %q, got %q", updatedWg.Name, updatedWidget.Name)
	}

	if updatedWidget.Weight != updatedWg.Weight {
		t.Errorf("expected widget.weight %d, got %d", updatedWg.Weight, updatedWidget.Weight)
	}

	if !updatedWidget.CreatedAt.Equal(createdWidget.CreatedAt) {
		t.Errorf("expected widget.createdat %q, got %q", createdWidget.CreatedAt, updatedWidget.CreatedAt)
	}

	if !updatedWidget.UpdatedAt.After(createdWidget.UpdatedAt) {
		t.Errorf("updatedat didn't changed, was: %q, now: %q", updatedWidget.UpdatedAt, createdWidget.UpdatedAt)
	}
}

func TestDelete(t *testing.T) {
	restoreDB(t)

	widgetRepo := postgresadapter.NewWidgetRepository(pgPool)

	wg := widget.Widget{
		Name:   "widget1",
		Weight: 1,
	}

	createdWidget, err := widgetRepo.Create(t.Context(), wg)
	if err != nil {
		t.Fatalf("unexpected error while creating widget: %v", err)
	}

	err = widgetRepo.Delete(t.Context(), createdWidget.ID)
	if err != nil {
		t.Fatalf("unexpected error while deleting widget: %v", err)
	}

	_, err = widgetRepo.Get(t.Context(), createdWidget.ID)
	if !errors.Is(err, apperror.ErrNotFound) {
		t.Errorf("expected notfound error, got: %v", err)
	}
}

func TestList(t *testing.T) {
	restoreDB(t)

	widgetRepo := postgresadapter.NewWidgetRepository(pgPool)

	widget1 := widget.Widget{
		Name:   "widget1",
		Weight: 1,
	}

	widget2 := widget.Widget{
		Name:   "widget2",
		Weight: 100,
	}

	_, err := widgetRepo.Create(t.Context(), widget1)
	if err != nil {
		t.Fatalf("unexpected error while creating widget: %v", err)
	}

	_, err = widgetRepo.Create(t.Context(), widget2)
	if err != nil {
		t.Fatalf("unexpected error while creating widget: %v", err)
	}

	widgetsList, err := widgetRepo.List(t.Context())
	if err != nil {
		t.Fatalf("unexpected error while fetching widgets list: %v", err)
	}

	if len(widgetsList) != 2 {
		t.Errorf("expected len: 2, got: %d", len(widgetsList))
	}

	_, err = pgPool.Exec(t.Context(), "TRUNCATE TABLE widgets;")
	if err != nil {
		t.Errorf("unexpected error while truncating: %v", err)
	}

	widgetsList, err = widgetRepo.List(t.Context())
	if err != nil {
		t.Errorf("unexpected error while fetching widgets list: %v", err)
	}

	if widgetsList == nil {
		t.Error("expected zero-len slice, got nil")
	}
}

func TestConstraint(t *testing.T) {
	restoreDB(t)

	widgetRepo := postgresadapter.NewWidgetRepository(pgPool)

	wg := widget.Widget{
		Name:   "w1",
		Weight: 1,
	}

	_, err := widgetRepo.Create(t.Context(), wg)
	if err == nil {
		t.Fatal("expected check-constraint violation, got success")
	}

	if errors.Is(err, apperror.ErrInvalid) || errors.Is(err, apperror.ErrNotFound) {
		t.Fatalf("constraint violations must pass through unmapped, got mapped kind: %v", err)
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" {
		t.Fatalf("expected pg check violation 23514 in chain, got: %v", err)
	}

}
