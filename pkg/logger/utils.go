package logger

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

func attachContextFields[T ChainableLogContext[T]](logCtx T, args ...any) T {
	if len(args) == 0 || len(args)%2 != 0 {
		return logCtx
	}

	for i := 0; i < len(args)-1; i += 2 {
		_key := args[i]
		_value := args[i+1]

		key, ok := _key.(string)
		if !ok {
			panic(fmt.Errorf("Log argument key should always be a string, got: %T", key))
		}

		switch val := _value.(type) {
		case string:
			logCtx = logCtx.Str(key, val)
		case int:
			logCtx = logCtx.Int(key, val)
		case float64:
			logCtx = logCtx.Float64(key, val)
		case error:
			logCtx = logCtx.Str(key, val.Error())
		default:
			logCtx = logCtx.Str(key, fmt.Sprintf("%+v", val))
		}
	}

	return logCtx
}

var CorrelationIDKey = "X-Correlation-ID"

func CorrelationIDFromContext(ctx context.Context) string {
	if v := ctx.Value(CorrelationIDKey); v != nil {
		return v.(string)
	}
	return ""
}

func CtxWithCorrelationID(ctx context.Context, id string) context.Context {
	if id == "" {
		id = uuid.New().String()
	}
	return context.WithValue(ctx, CorrelationIDKey, id)
}
