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

func TestSequenceStepFromDB(t *testing.T) {
	now := time.Now()
	step := &models.SequenceStep{
		ID:                    uuid.New(),
		EmailSubject:          "Test Subject",
		EmailContent:          "Test Content",
		DaysAfterPreviousStep: 1,
		CreatedAt:             pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:             pgtype.Timestamptz{Time: now, Valid: true},
	}

	result := SequenceStepFromDB(step)

	assert.Equal(t, step.ID, result.Id)
	assert.Equal(t, step.EmailSubject, result.EmailSubject)
	assert.Equal(t, step.EmailContent, result.EmailContent)
	assert.Equal(t, int(step.DaysAfterPreviousStep), result.DaysAfterPreviousStep)
	assert.Equal(t, &step.CreatedAt.Time, result.CreatedAt)
	assert.Equal(t, &step.UpdatedAt.Time, result.UpdatedAt)
}

func TestUpdateSequenceStep(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockSequenceService(ctrl)
	handler := &StrictHandler{svc: mockService}
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		sequenceID := uuid.New()
		stepID := uuid.New()
		now := time.Now()

		emailSubject := "Updated Subject"
		emailContent := "Updated Content"

		request := openapi.UpdateSequenceStepRequestObject{
			SequenceId: sequenceID.String(),
			StepId:     stepID.String(),
			Body: &openapi.UpdateSequenceStepInput{
				EmailSubject: &emailSubject,
				EmailContent: &emailContent,
			},
		}

		expectedStep := &models.SequenceStep{
			ID:                    stepID,
			EmailSubject:          emailSubject,
			EmailContent:          emailContent,
			DaysAfterPreviousStep: 1,
			CreatedAt:             pgtype.Timestamptz{Time: now, Valid: true},
			UpdatedAt:             pgtype.Timestamptz{Time: now, Valid: true},
		}

		mockService.EXPECT().
			UpdateSequenceStep(ctx, sequenceID, stepID, &emailSubject, &emailContent).
			Return(expectedStep, nil)

		response, err := handler.UpdateSequenceStep(ctx, request)
		assert.NoError(t, err)

		result := response.(openapi.UpdateSequenceStep200JSONResponse)
		assert.Equal(t, expectedStep.ID, result.Id)
		assert.Equal(t, expectedStep.EmailSubject, result.EmailSubject)
		assert.Equal(t, expectedStep.EmailContent, result.EmailContent)
	})

	t.Run("handles invalid sequence UUID", func(t *testing.T) {
		emailSubject := "Updated Subject"
		emailContent := "Updated Content"

		request := openapi.UpdateSequenceStepRequestObject{
			SequenceId: "invalid-uuid",
			StepId:     uuid.New().String(),
			Body: &openapi.UpdateSequenceStepInput{
				EmailSubject: &emailSubject,
				EmailContent: &emailContent,
			},
		}

		response, err := handler.UpdateSequenceStep(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid sequence ID")
	})

	t.Run("handles invalid step UUID", func(t *testing.T) {
		emailSubject := "Updated Subject"
		emailContent := "Updated Content"

		request := openapi.UpdateSequenceStepRequestObject{
			SequenceId: uuid.New().String(),
			StepId:     "invalid-uuid",
			Body: &openapi.UpdateSequenceStepInput{
				EmailSubject: &emailSubject,
				EmailContent: &emailContent,
			},
		}

		response, err := handler.UpdateSequenceStep(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid step ID")
	})

	t.Run("handles not found error", func(t *testing.T) {
		sequenceID := uuid.New()
		stepID := uuid.New()

		emailSubject := "Updated Subject"
		emailContent := "Updated Content"

		request := openapi.UpdateSequenceStepRequestObject{
			SequenceId: sequenceID.String(),
			StepId:     stepID.String(),
			Body: &openapi.UpdateSequenceStepInput{
				EmailSubject: &emailSubject,
				EmailContent: &emailContent,
			},
		}

		mockService.EXPECT().
			UpdateSequenceStep(ctx, sequenceID, stepID, &emailSubject, &emailContent).
			Return(nil, sql.ErrNoRows)

		response, err := handler.UpdateSequenceStep(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Sequence step not found")
	})

	t.Run("handles internal error", func(t *testing.T) {
		sequenceID := uuid.New()
		stepID := uuid.New()

		emailSubject := "Updated Subject"
		emailContent := "Updated Content"

		request := openapi.UpdateSequenceStepRequestObject{
			SequenceId: sequenceID.String(),
			StepId:     stepID.String(),
			Body: &openapi.UpdateSequenceStepInput{
				EmailSubject: &emailSubject,
				EmailContent: &emailContent,
			},
		}

		mockService.EXPECT().
			UpdateSequenceStep(ctx, sequenceID, stepID, &emailSubject, &emailContent).
			Return(nil, assert.AnError)

		response, err := handler.UpdateSequenceStep(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to update sequence step")
	})
}

func TestDeleteSequenceStep(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := NewMockSequenceService(ctrl)
	handler := &StrictHandler{svc: mockService}
	ctx := context.Background()

	t.Run("successful deletion", func(t *testing.T) {
		sequenceID := uuid.New()
		stepID := uuid.New()

		request := openapi.DeleteSequenceStepRequestObject{
			SequenceId: sequenceID.String(),
			StepId:     stepID.String(),
		}

		mockService.EXPECT().
			DeleteSequenceStep(ctx, sequenceID, stepID).
			Return(nil)

		response, err := handler.DeleteSequenceStep(ctx, request)
		assert.NoError(t, err)
		assert.IsType(t, openapi.DeleteSequenceStep204Response{}, response)
	})

	t.Run("handles invalid sequence UUID", func(t *testing.T) {
		request := openapi.DeleteSequenceStepRequestObject{
			SequenceId: "invalid-uuid",
			StepId:     uuid.New().String(),
		}

		response, err := handler.DeleteSequenceStep(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid sequence ID")
	})

	t.Run("handles invalid step UUID", func(t *testing.T) {
		request := openapi.DeleteSequenceStepRequestObject{
			SequenceId: uuid.New().String(),
			StepId:     "invalid-uuid",
		}

		response, err := handler.DeleteSequenceStep(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid step ID")
	})

	t.Run("handles internal error", func(t *testing.T) {
		sequenceID := uuid.New()
		stepID := uuid.New()

		request := openapi.DeleteSequenceStepRequestObject{
			SequenceId: sequenceID.String(),
			StepId:     stepID.String(),
		}

		mockService.EXPECT().
			DeleteSequenceStep(ctx, sequenceID, stepID).
			Return(assert.AnError)

		response, err := handler.DeleteSequenceStep(ctx, request)
		assert.Nil(t, response)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Failed to delete sequence step")
	})
}
