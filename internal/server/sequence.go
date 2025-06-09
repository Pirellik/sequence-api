package server

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/pirellik/sequence-api/internal/db/models"
	"github.com/pirellik/sequence-api/internal/openapi"
	"github.com/samber/lo"
)

func SequenceFromDB(sequence *models.Sequence, steps []*models.SequenceStep) openapi.Sequence {
	return openapi.Sequence{
		Id:                   sequence.ID,
		Name:                 sequence.Name,
		OpenTrackingEnabled:  sequence.OpenTrackingEnabled,
		ClickTrackingEnabled: sequence.ClickTrackingEnabled,
		Steps: lo.Map(steps, func(step *models.SequenceStep, _ int) openapi.SequenceStep {
			return SequenceStepFromDB(step)
		}),
		CreatedAt: &sequence.CreatedAt.Time,
		UpdatedAt: &sequence.UpdatedAt.Time,
	}
}

func (s *StrictHandler) CreateSequence(ctx context.Context, request openapi.CreateSequenceRequestObject) (openapi.CreateSequenceResponseObject, error) {
	sequence := models.Sequence{
		Name:                 request.Body.Name,
		OpenTrackingEnabled:  request.Body.OpenTrackingEnabled,
		ClickTrackingEnabled: request.Body.ClickTrackingEnabled,
	}

	steps := lo.Map(request.Body.Steps, func(step openapi.SequenceStep, _ int) *models.SequenceStep {
		return &models.SequenceStep{
			EmailSubject:          step.EmailSubject,
			EmailContent:          step.EmailContent,
			DaysAfterPreviousStep: int32(step.DaysAfterPreviousStep),
		}
	})

	createdSequence, createdSteps, err := s.svc.CreateSequence(ctx, &sequence, steps)
	if err != nil {
		return nil, ErrInternal("Failed to create sequence")
	}

	return openapi.CreateSequence201JSONResponse(SequenceFromDB(createdSequence, createdSteps)), nil
}

func (s *StrictHandler) UpdateSequence(ctx context.Context, request openapi.UpdateSequenceRequestObject) (openapi.UpdateSequenceResponseObject, error) {
	id, err := uuid.Parse(request.Id)
	if err != nil {
		return nil, ErrBadRequest("Invalid sequence ID")
	}

	updatedSequence, updatedSteps, err := s.svc.UpdateSequence(ctx, id, request.Body.OpenTrackingEnabled, request.Body.ClickTrackingEnabled)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound("Sequence not found")
		}
		return nil, ErrInternal("Failed to create sequence")
	}

	return openapi.UpdateSequence200JSONResponse(SequenceFromDB(updatedSequence, updatedSteps)), nil
}
