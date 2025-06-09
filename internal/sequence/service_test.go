package sequence

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/go-testfixtures/testfixtures/v3"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pirellik/sequence-api/internal/db"
	"github.com/pirellik/sequence-api/internal/db/models"
	"github.com/pirellik/sequence-api/pkg/pointer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testDB struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
}

func setupTestDB(t *testing.T) *testDB {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:17.5",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	connString, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	err = db.MigrateUp(connString)
	require.NoError(t, err)

	// Setup testfixtures
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	defer conn.Release()

	sqlDB, err := sql.Open("postgres", connString)
	require.NoError(t, err)

	fixtures, err := testfixtures.New(
		testfixtures.Database(sqlDB),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory("../db/fixtures"),
	)
	require.NoError(t, err)

	err = fixtures.Load()
	require.NoError(t, err)

	return &testDB{
		container: pgContainer,
		pool:      pool,
	}
}

func (db *testDB) cleanup(t *testing.T) {
	db.pool.Close()
	err := db.container.Terminate(context.Background())
	require.NoError(t, err)
}

func TestCreateSequence(t *testing.T) {
	db := setupTestDB(t)
	defer db.cleanup(t)

	service := NewService(db.pool)
	ctx := context.Background()

	t.Run("valid sequence without steps", func(t *testing.T) {
		sequence, steps, err := service.CreateSequence(ctx, &models.Sequence{
			Name:                 "New Test Sequence",
			OpenTrackingEnabled:  true,
			ClickTrackingEnabled: true,
		}, nil)
		require.NoError(t, err)
		require.NotNil(t, sequence)
		assert.Equal(t, "New Test Sequence", sequence.Name)
		assert.True(t, sequence.OpenTrackingEnabled)
		assert.True(t, sequence.ClickTrackingEnabled)
		assert.Empty(t, steps)
	})

	t.Run("valid sequence with steps", func(t *testing.T) {
		sequence, steps, err := service.CreateSequence(ctx, &models.Sequence{
			Name:                 "Another New Test Sequence",
			OpenTrackingEnabled:  true,
			ClickTrackingEnabled: true,
		}, []*models.SequenceStep{
			{
				EmailSubject:          "Step 1",
				EmailContent:          "Content 1",
				DaysAfterPreviousStep: 1,
			},
			{
				EmailSubject:          "Step 2",
				EmailContent:          "Content 2",
				DaysAfterPreviousStep: 2,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, sequence)
		assert.Equal(t, "Another New Test Sequence", sequence.Name)
		assert.True(t, sequence.OpenTrackingEnabled)
		assert.True(t, sequence.ClickTrackingEnabled)
		assert.Len(t, steps, 2)
		assert.Equal(t, "Step 1", steps[0].EmailSubject)
		assert.Equal(t, "Step 2", steps[1].EmailSubject)
		assert.Equal(t, "Content 1", steps[0].EmailContent)
		assert.Equal(t, "Content 2", steps[1].EmailContent)
	})
}

func TestUpdateSequence(t *testing.T) {
	db := setupTestDB(t)
	defer db.cleanup(t)

	service := NewService(db.pool)
	ctx := context.Background()

	sequenceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	tests := []struct {
		name                 string
		sequenceID           uuid.UUID
		openTrackingEnabled  *bool
		clickTrackingEnabled *bool
		wantErr              bool
	}{
		{
			name:                 "update open tracking",
			sequenceID:           sequenceID,
			openTrackingEnabled:  pointer.To(false),
			clickTrackingEnabled: nil,
			wantErr:              false,
		},
		{
			name:                 "update click tracking",
			sequenceID:           sequenceID,
			openTrackingEnabled:  nil,
			clickTrackingEnabled: pointer.To(false),
			wantErr:              false,
		},
		{
			name:                 "update both",
			sequenceID:           sequenceID,
			openTrackingEnabled:  pointer.To(false),
			clickTrackingEnabled: pointer.To(false),
			wantErr:              false,
		},
		{
			name:                 "not found",
			sequenceID:           uuid.New(),
			openTrackingEnabled:  nil,
			clickTrackingEnabled: nil,
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, _, err := service.UpdateSequence(ctx, tt.sequenceID, tt.openTrackingEnabled, tt.clickTrackingEnabled)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, updated)

			if tt.openTrackingEnabled != nil {
				assert.Equal(t, *tt.openTrackingEnabled, updated.OpenTrackingEnabled)
			}
			if tt.clickTrackingEnabled != nil {
				assert.Equal(t, *tt.clickTrackingEnabled, updated.ClickTrackingEnabled)
			}
		})
	}
}

func TestUpdateSequenceStep(t *testing.T) {
	db := setupTestDB(t)
	defer db.cleanup(t)

	service := NewService(db.pool)
	ctx := context.Background()

	sequenceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	stepID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	tests := []struct {
		name         string
		stepID       uuid.UUID
		emailSubject *string
		emailContent *string
		wantErr      bool
	}{
		{
			name:         "update subject",
			stepID:       stepID,
			emailSubject: pointer.To("New Subject"),
			emailContent: nil,
			wantErr:      false,
		},
		{
			name:         "update content",
			stepID:       stepID,
			emailSubject: nil,
			emailContent: pointer.To("New Content"),
			wantErr:      false,
		},
		{
			name:         "not found",
			stepID:       uuid.New(),
			emailSubject: nil,
			emailContent: nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updated, err := service.UpdateSequenceStep(ctx, sequenceID, tt.stepID, tt.emailSubject, tt.emailContent)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, updated)

			if tt.emailSubject != nil {
				assert.Equal(t, *tt.emailSubject, updated.EmailSubject)
			}
			if tt.emailContent != nil {
				assert.Equal(t, *tt.emailContent, updated.EmailContent)
			}
		})
	}
}

func TestDeleteSequenceStep(t *testing.T) {
	db := setupTestDB(t)
	defer db.cleanup(t)

	service := NewService(db.pool)
	ctx := context.Background()

	sequenceID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	stepID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	err := service.DeleteSequenceStep(ctx, sequenceID, stepID)
	require.NoError(t, err)

	// Verify step was deleted
	q := models.New(db.pool)
	_, err = q.GetSequenceStepByID(ctx, stepID)
	assert.Error(t, err) // Should get an error as the step no longer exists
}
