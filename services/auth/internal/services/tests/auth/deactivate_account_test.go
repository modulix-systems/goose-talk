package auth_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/modulix-systems/goose-talk/internal/services/auth"
	"github.com/modulix-systems/goose-talk/tests/suite"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeactivateAccSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	mockUser := suite.MockUser()
	userId := strconv.Itoa(mockUser.ID)
	authSuite.mockUsersRepo.EXPECT().UpdateIsActiveById(ctx, userId, false).Return(mockUser, nil)
	authSuite.mockMailSender.EXPECT().SendAccDeactivationEmail(ctx, mockUser.Email).Return(nil)

	err := authSuite.service.DeactivateAccount(ctx, userId)

	assert.NoError(t, err)
}

func TestDeactivateAccUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	authSuite := NewAuthTestSuite(ctrl)
	ctx := context.Background()
	userId := strconv.Itoa(gofakeit.Number(1, 100))
	authSuite.mockUsersRepo.EXPECT().UpdateIsActiveById(ctx, userId, false).Return(nil, storage.ErrNotFound)

	err := authSuite.service.DeactivateAccount(ctx, userId)

	assert.ErrorIs(t, err, auth.ErrUserNotFound)
}
