package auth_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeactivateSessionSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockSessionId := gofakeit.UUID()
	authSuite.mockSessionsRepo.EXPECT().UpdateIsActiveById(ctx, mockSessionId, false).Return(nil)

	err := authSuite.service.DeactivateSession(ctx, mockSessionId)
	assert.NoError(t, err)
}

func TestDeactivateSessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockSessionId := gofakeit.UUID()
	authSuite.mockSessionsRepo.EXPECT().UpdateIsActiveById(ctx, mockSessionId, false).Return(storage.ErrNotFound)

	err := authSuite.service.DeactivateSession(ctx, mockSessionId)
	assert.ErrorIs(t, err, auth.ErrSessionNotFound)
}
