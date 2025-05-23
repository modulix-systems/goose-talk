package auth_test

// import (
// 	"context"
// 	"testing"
// 	"time"
//
// 	"github.com/brianvoe/gofakeit/v7"
// 	"github.com/modulix-systems/goose-talk/internal/entity"
// 	"github.com/modulix-systems/goose-talk/tests/suite/helpers"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"
// )
//
// func TestExportLoginTokenSuccess(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	authSuite := NewAuthTestSuite(ctrl)
// 	ctx := context.Background()
// 	mockSessionId := gofakeit.UUID()
// 	mockLoginToken := helpers.MockLoginToken(authSuite.tokenTTL)
// 	authSuite.mockLoginTokenRepo.EXPECT().Insert(ctx, mockLoginToken)
//
// 	token, err := authSuite.service.ExportLoginToken(ctx, mockSessionId)
// 	assert.NoError(t, err)
// 	assert.Equal(t, token.SessionId, mockSessionId)
// 	assert.Empty(t, token.AuthSessionId)
// 	assert.WithinDuration(t, time.Now().Add(authSuite.tokenTTL), token.ExpiresAt, time.Second)
// }
//
// func TestExportLoginTokenApprovedSuccess(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	authSuite := NewAuthTestSuite(ctrl)
// 	ctx := context.Background()
// 	mockSessionId := gofakeit.UUID()
//
// 	token, err := authSuite.service.ExportLoginToken(ctx, mockSessionId)
// 	assert.NoError(t, err)
// 	assert.Equal(t, token.SessionId, mockSessionId)
// 	assert.Empty(t, token.AuthSessionId)
// 	assert.WithinDuration(t, time.Now().Add(authSuite.tokenTTL), token.ExpiresAt, time.Second)
// }
//
