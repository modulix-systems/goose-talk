package utils

import (
	"context"

	"github.com/modulix-systems/goose-talk/logger"
	"google.golang.org/grpc/metadata"
)

func GetCorrelationIdFromGrpcCtx(ctx context.Context) string {
	meta, exists := metadata.FromIncomingContext(ctx)
	if !exists {
		return ""
	}

	correlationId := meta.Get(logger.CorrelationIDKey)
	if len(correlationId) > 0 {
		return correlationId[0]
	}

	return ""
}
