package describer

import (
	"context"
	"github.com/opengovern/og-util/pkg/describe/enums"
	"go.uber.org/zap"
	"oras.land/oras-go/v2/registry/remote"
)

var (
	triggerTypeKey string = "trigger_type"
)

func WithTriggerType(ctx context.Context, tt enums.DescribeTriggerType) context.Context {
	return context.WithValue(ctx, triggerTypeKey, tt)
}

func GetTriggerTypeFromContext(ctx context.Context) enums.DescribeTriggerType {
	tt, ok := ctx.Value(triggerTypeKey).(enums.DescribeTriggerType)
	if !ok {
		return ""
	}
	return tt
}

func GetParameterFromContext(ctx context.Context, key string) any {
	return ctx.Value(key)
}

func WithLogger(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, "logger", logger)
}

func GetLoggerFromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value("logger").(*zap.Logger)
	if !ok {
		return zap.NewNop()
	}
	return logger
}

func WithOrasClient(ctx context.Context, client remote.Client) context.Context {
	return context.WithValue(ctx, "oras_client", client)
}

func GetOrasClientFromContext(ctx context.Context) remote.Client {
	client, ok := ctx.Value("oras_client").(remote.Client)
	if !ok {
		return nil
	}
	return client
}

func WithRegistry(ctx context.Context, reg string) context.Context {
	return context.WithValue(ctx, "registry", reg)
}

func GetRegistryFromContext(ctx context.Context) string {
	reg, ok := ctx.Value("registry").(string)
	if !ok {
		return ""
	}
	return reg
}

func WithOwner(ctx context.Context, owner string) context.Context {
	return context.WithValue(ctx, "owner", owner)
}

func GetOwnerFromContext(ctx context.Context) string {
	owner, ok := ctx.Value("owner").(string)
	if !ok {
		return ""
	}
	return owner
}
