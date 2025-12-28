package postgres_test

import (
	"context"
	"testing"

	"github.com/modulix-systems/goose-talk/internal/config"
	"github.com/modulix-systems/goose-talk/pkg/postgres"
	"github.com/modulix-systems/goose-talk/tests/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetQueryableWithTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockConnProvider := mocks.NewMockSupportsAcquire(ctrl)
	expectedTx := mocks.NewMockQueryable(ctrl)
	ctx := context.WithValue(context.Background(), config.TRANSACTION_CTX_KEY, expectedTx)
	queryable, err := postgres.GetQueryable(ctx, mockConnProvider)
	assert.NoError(t, err)
	assert.Equal(t, expectedTx, queryable)
}

func TestGetQueryableNewConn(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	expectedConn := mocks.NewMockQueryable(ctrl)
	mockConnProvider := mocks.NewMockSupportsAcquire(ctrl)
	mockConnProvider.EXPECT().Acquire(context.Background()).Return(expectedConn, nil)
	queryable, err := postgres.GetQueryable(ctx, mockConnProvider)
	assert.NoError(t, err)
	assert.Equal(t, expectedConn, queryable)
}
