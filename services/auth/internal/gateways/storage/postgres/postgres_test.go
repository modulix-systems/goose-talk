package postgres_repos_test

import (
	"context"
	"testing"

	postgres_repos "github.com/modulix-systems/goose-talk/internal/gateways/storage/postgres"
	"github.com/modulix-systems/goose-talk/internal/services"
	"github.com/modulix-systems/goose-talk/tests/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetQueryableWithTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockConnProvider := mocks.NewMockSupportsAcquire(ctrl)
	expectedTx := mocks.NewMockQueryable(ctrl)
	ctx := context.WithValue(context.Background(), services.TransactionCtxKey, expectedTx)
	queryable, err := postgres_repos.GetQueryable(ctx, mockConnProvider)
	assert.NoError(t, err)
	assert.Equal(t, expectedTx, queryable)
}

func TestGetQueryableNewConn(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	expectedConn := mocks.NewMockQueryable(ctrl)
	mockConnProvider := mocks.NewMockSupportsAcquire(ctrl)
	mockConnProvider.EXPECT().Acquire(context.Background()).Return(expectedConn, nil)
	queryable, err := postgres_repos.GetQueryable(ctx, mockConnProvider)
	assert.NoError(t, err)
	assert.Equal(t, expectedConn, queryable)
}
