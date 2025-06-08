package services

import (
	"context"
	"fmt"

	"github.com/looplab/fsm"
	"github.com/modulix-systems/goose-talk/internal/gateways"
)

type contextKey string

const TransactionCtxKey contextKey = "dbTrx"

func SetTransaction(ctx context.Context, tx any) context.Context {
	return context.WithValue(ctx, contextKey(TransactionCtxKey), tx)
}

type usecaseState string

const (
	UsecaseStarted usecaseState = "started"
	UsecaseFailed  usecaseState = "failed"
	UsecaseSucceed usecaseState = "succeed"
)

type UsecaseStateMachine struct {
	fsm         *fsm.FSM
	transaction gateways.Transaction
}

func NewUsecaseState(transaction gateways.Transaction) *UsecaseStateMachine {
	return &UsecaseStateMachine{
		fsm: fsm.NewFSM(
			string(UsecaseStarted),
			fsm.Events{
				{Name: string(UsecaseStarted), Src: []string{string(UsecaseSucceed), string(UsecaseFailed)}, Dst: string(UsecaseStarted)},
				{Name: string(UsecaseFailed), Src: []string{string(UsecaseStarted)}, Dst: string(UsecaseFailed)},
				{Name: string(UsecaseSucceed), Src: []string{string(UsecaseStarted)}, Dst: string(UsecaseSucceed)},
			},
			fsm.Callbacks{
				string(UsecaseSucceed): func(ctx context.Context, e *fsm.Event) {
					if err := transaction.Commit(ctx); err != nil {
						fmt.Errorf("failed to commit transaction: %w", err)
					}
				},
				string(UsecaseFailed): func(ctx context.Context, e *fsm.Event) {
					if err := transaction.Rollback(ctx); err != nil {
						fmt.Errorf("failed to rollback transaction: %w", err)
					}
				},
			}),
		transaction: transaction,
	}
}
