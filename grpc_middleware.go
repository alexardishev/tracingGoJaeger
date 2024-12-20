package tracing_lib

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	// "go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Интерсептор, добавляющий трейсинг для gRPC-сервера. По сути добавляем его в gRPC и все методы работают.
func UnaryServerInterceptor(tracerName string) grpc.UnaryServerInterceptor {
	tracer := otel.Tracer(tracerName)

	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp any, err error) {
		//Извлекаю данные из контекста метадату
		md, _ := metadata.FromIncomingContext(ctx)
		//Извлекаю из метадаты данные по трейсингу и обогащаю текущий контекст
		parentCtx := otel.GetTextMapPropagator().Extract(ctx, metadataCarrier(md))
		//Стартую новый спан, если есть родительский, тогда привязываю к родительскому новый спан
		ctx, span := tracer.Start(parentCtx, info.FullMethod)
		defer span.End()

		span.SetAttributes(
			attribute.String("rpc.method", info.FullMethod),
		)
		for key, values := range md {
			span.SetAttributes(attribute.StringSlice("rpc.metadata."+key, values))
		}
		span.SetAttributes(
			attribute.String("rpc.request.body", fmt.Sprintf("%+v", req)),
		)

		resp, err = handler(ctx, req)

		if err != nil {
			span.SetAttributes(attribute.String("rpc.error", err.Error()))
		}

		return resp, err
	}
}

// metadataCarrier адаптирует метаданные gRPC для OpenTelemetry.
type metadataCarrier metadata.MD

func (mc metadataCarrier) Get(key string) string {
	values := metadata.MD(mc).Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (mc metadataCarrier) Set(key, value string) {
	metadata.MD(mc).Set(key, value)
}

func (mc metadataCarrier) Keys() []string {
	var keys []string
	for key := range mc {
		keys = append(keys, key)
	}
	return keys
}
