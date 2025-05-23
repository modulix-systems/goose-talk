package auth_test

// import (
// 	"context"
// 	"testing"
//
// 	"github.com/brianvoe/gofakeit/v7"
// 	"github.com/stretchr/testify/assert"
// 	"go.uber.org/mock/gomock"
// )
//
// func TestExportLoginTokenSuccess(t *testing.T) {
// 	ctrl := gomock.NewController(t)
// 	authSuite := NewAuthTestSuite(ctrl)
// 	ctx := context.Background()
// 	mockConnId := gofakeit.UUID()
//
// 	token, err := authSuite.service.ExportLoginToken(ctx, mockConnId)
// 	assert.NoError(t, err)
// 	assert.Equal(t, token.ConnId, mockConnId)
// }
