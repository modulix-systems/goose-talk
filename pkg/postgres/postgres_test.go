package postgres_test

import (
	"context"
	"testing"

	"github.com/modulix-systems/goose-talk/postgres"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const txKey = "pg-transaction"

func TestGetQueryableWithTx(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockConnProvider := NewMockAcquirable(ctrl)
	expectedTx := NewMockQueryable(ctrl)
	ctx := context.WithValue(context.Background(), txKey, expectedTx)
	queryable, err := postgres.GetQueryable(ctx, mockConnProvider, txKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedTx, queryable)
}

func TestGetQueryableNewConn(t *testing.T) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()
	expectedConn := NewMockQueryable(ctrl)
	mockConnProvider := NewMockAcquirable(ctrl)
	mockConnProvider.EXPECT().Acquire(context.Background()).Return(expectedConn, nil)
	queryable, err := postgres.GetQueryable(ctx, mockConnProvider, txKey)
	assert.NoError(t, err)
	assert.Equal(t, expectedConn, queryable)
}
