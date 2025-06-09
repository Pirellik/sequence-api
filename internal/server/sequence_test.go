package server

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pirellik/sequence-api/internal/db/models"
	"github.com/pirellik/sequence-api/internal/openapi"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSequenceFromDB(t *testing.T) {
	now := time.Now()
	sequence := &models.Sequence{
		ID:                   uuid.New(),
		Name:                 "Test Sequence",
		OpenTrackingEnabled:  true,
		ClickTrackingEnabled: false,
		CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
	}

	steps := []*models.SequenceStep{
		{
			ID:                    uuid.New(),
			EmailSubject:          "Test Subject",
			EmailContent:          "Test Content",
			DaysAfterPreviousStep: 1,
		},
	}

	result := SequenceFromDB(sequence, steps)

	assert.Equal(t, sequence.ID, result.Id)
	assert.Equal(t, sequence.Name, result.Name)
	assert.Equal(t, sequence.OpenTrackingEnabled, result.OpenTrackingEnabled)
	assert.Equal(t, sequence.ClickTrackingEnabled, result.ClickTrackingEnabled)
	assert.Equal(t, &sequence.CreatedAt.Time, result.CreatedAt)
	assert.Equal(t, &sequence.UpdatedAt.Time, result.UpdatedAt)
	assert.Len(t, result.Steps, 1)
	assert.Equal(t, steps[0].EmailSubject, result.Steps[0].EmailSubject)
	assert.Equal(t, steps[0].EmailContent, result.Steps[0].EmailContent)
	assert.Equal(t, int(steps[0].DaysAfterPreviousStep), result.Steps[0].DaysAfterPreviousStep)
}

func TestCreateSequence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockSequenceService(ctrl)
	handler := &StrictHandler{svc: mockService}
	ctx := context.Background()

	t.Run("successful creation with steps", func(t *testing.T) {
		request := openapi.CreateSequenceRequestObject{
			Body: &openapi.Sequence{
				Name:                 "Test Sequence",
				OpenTrackingEnabled:  true,
				ClickTrackingEnabled: false,
				Steps: []openapi.SequenceStep{
					{
						EmailSubject:          "Test Subject",
						EmailContent:          "Test Content",
						DaysAfterPreviousStep: 1,
					},
				},
			},
		}

		now := time.Now()
		expectedSequence := &models.Sequence{
			ID:                   uuid.New(),
			Name:                 request.Body.Name,
			OpenTrackingEnabled:  request.Body.OpenTrackingEnabled,
			ClickTrackingEnabled: request.Body.ClickTrackingEnabled,
			CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		}

		expectedSteps := []*models.SequenceStep{
			{
				ID:                    uuid.New(),
				EmailSubject:          "Test Subject",
				EmailContent:          "Test Content",
				DaysAfterPreviousStep: 1,
			},
		}

		mockService.EXPECT().
			CreateSequence(ctx, gomock.Any(), gomock.Any()).
			Return(expectedSequence, expectedSteps, nil)

		response, err := handler.CreateSequence(ctx, request)
		assert.NoError(t, err)

		result := response.(openapi.CreateSequence201JSONResponse)
		assert.Equal(t, expectedSequence.ID, result.Id)
		assert.Equal(t, expectedSequence.Name, result.Name)
		assert.Equal(t, expectedSequence.OpenTrackingEnabled, result.OpenTrackingEnabled)
		assert.Equal(t, expectedSequence.ClickTrackingEnabled, result.ClickTrackingEnabled)
		assert.Len(t, result.Steps, 1)
	})

	t.Run("successful creation without steps", func(t *testing.T) {
		request := openapi.CreateSequenceRequestObject{
			Body: &openapi.Sequence{
				Name:                 "Test Sequence",
				OpenTrackingEnabled:  true,
				ClickTrackingEnabled: false,
				Steps:                []openapi.SequenceStep{},
			},
		}

		now := time.Now()
		expectedSequence := &models.Sequence{
			ID:                   uuid.New(),
			Name:                 request.Body.Name,
			OpenTrackingEnabled:  request.Body.OpenTrackingEnabled,
			ClickTrackingEnabled: request.Body.ClickTrackingEnabled,
			CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		}

		mockService.EXPECT().
			CreateSequence(ctx, gomock.Any(), gomock.Any()).
			Return(expectedSequence, []*models.SequenceStep{}, nil)

		response, err := handler.CreateSequence(ctx, request)
		assert.NoError(t, err)

		result := response.(openapi.CreateSequence201JSONResponse)
		assert.Equal(t, expectedSequence.ID, result.Id)
		assert.Equal(t, expectedSequence.Name, result.Name)
		assert.Equal(t, expectedSequence.OpenTrackingEnabled, result.OpenTrackingEnabled)
		assert.Equal(t, expectedSequence.ClickTrackingEnabled, result.ClickTrackingEnabled)
		assert.Empty(t, result.Steps)
	})

	t.Run("handles service error", func(t *testing.T) {
		request := openapi.CreateSequenceRequestObject{
			Body: &openapi.Sequence{
				Name:                 "Test Sequence",
				OpenTrackingEnabled:  true,
				ClickTrackingEnabled: false,
				Steps:                []openapi.SequenceStep{},
			},
		}

		mockService.EXPECT().
			CreateSequence(ctx, gomock.Any(), gomock.Any()).
			Return(nil, nil, assert.AnError)

		response, err := handler.CreateSequence(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to create sequence")
	})
}

func TestUpdateSequence(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockSequenceService(ctrl)
	handler := &StrictHandler{svc: mockService}
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		sequenceID := uuid.New()
		openTracking := true
		clickTracking := false

		request := openapi.UpdateSequenceRequestObject{
			Id: sequenceID.String(),
			Body: &openapi.UpdateSequenceInput{
				OpenTrackingEnabled:  &openTracking,
				ClickTrackingEnabled: &clickTracking,
			},
		}

		now := time.Now()
		expectedSequence := &models.Sequence{
			ID:                   sequenceID,
			Name:                 "Test Sequence",
			OpenTrackingEnabled:  openTracking,
			ClickTrackingEnabled: clickTracking,
			CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		}

		expectedSteps := []*models.SequenceStep{
			{
				ID:                    uuid.New(),
				EmailSubject:          "Test Subject",
				EmailContent:          "Test Content",
				DaysAfterPreviousStep: 1,
			},
		}

		mockService.EXPECT().
			UpdateSequence(ctx, sequenceID, &openTracking, &clickTracking).
			Return(expectedSequence, expectedSteps, nil)

		response, err := handler.UpdateSequence(ctx, request)
		assert.NoError(t, err)

		result := response.(openapi.UpdateSequence200JSONResponse)
		assert.Equal(t, expectedSequence.ID, result.Id)
		assert.Equal(t, expectedSequence.Name, result.Name)
		assert.Equal(t, expectedSequence.OpenTrackingEnabled, result.OpenTrackingEnabled)
		assert.Equal(t, expectedSequence.ClickTrackingEnabled, result.ClickTrackingEnabled)
		assert.Len(t, result.Steps, 1)
	})

	t.Run("handles invalid UUID", func(t *testing.T) {
		request := openapi.UpdateSequenceRequestObject{
			Id:   "invalid-uuid",
			Body: &openapi.UpdateSequenceInput{},
		}

		response, err := handler.UpdateSequence(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid sequence ID")
	})

	t.Run("handles not found error", func(t *testing.T) {
		sequenceID := uuid.New()
		openTracking := true
		clickTracking := false

		request := openapi.UpdateSequenceRequestObject{
			Id: sequenceID.String(),
			Body: &openapi.UpdateSequenceInput{
				OpenTrackingEnabled:  &openTracking,
				ClickTrackingEnabled: &clickTracking,
			},
		}

		mockService.EXPECT().
			UpdateSequence(ctx, sequenceID, &openTracking, &clickTracking).
			Return(nil, nil, sql.ErrNoRows)

		response, err := handler.UpdateSequence(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Sequence not found")
	})

	t.Run("handles partial update", func(t *testing.T) {
		sequenceID := uuid.New()
		openTracking := true

		request := openapi.UpdateSequenceRequestObject{
			Id: sequenceID.String(),
			Body: &openapi.UpdateSequenceInput{
				OpenTrackingEnabled: &openTracking,
			},
		}

		now := time.Now()
		expectedSequence := &models.Sequence{
			ID:                   sequenceID,
			Name:                 "Test Sequence",
			OpenTrackingEnabled:  openTracking,
			ClickTrackingEnabled: false,
			CreatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:            pgtype.Timestamptz{Time: now, Valid: true},
		}

		mockService.EXPECT().
			UpdateSequence(ctx, sequenceID, &openTracking, nil).
			Return(expectedSequence, []*models.SequenceStep{}, nil)

		response, err := handler.UpdateSequence(ctx, request)
		assert.NoError(t, err)

		result := response.(openapi.UpdateSequence200JSONResponse)
		assert.Equal(t, expectedSequence.ID, result.Id)
		assert.Equal(t, expectedSequence.Name, result.Name)
		assert.Equal(t, expectedSequence.OpenTrackingEnabled, result.OpenTrackingEnabled)
		assert.Equal(t, expectedSequence.ClickTrackingEnabled, result.ClickTrackingEnabled)
		assert.Empty(t, result.Steps)
	})
}
