package tracing_lib

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

// TracerProvider глобальный провайдер трейсинга
var TracerProvider *sdktrace.TracerProvider

// InitTracing инициализирует OpenTelemetry с OTLP
func InitTracing(serviceName, otlpEndpoint string) error {
	// Создаём OTLP экспортёр
	exp, err := otlptracegrpc.New(
		context.Background(),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
		otlptracegrpc.WithInsecure(), // Jaeger OTLP по умолчанию без TLS
	)
	if err != nil {
		return err
	}

	// Создаём ресурс с именем сервиса
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName), // Используем корректное создание KeyValue
		),
	)
	if err != nil {
		return err
	}

	// Создаём провайдер трейсинга
	TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),  // Отправляем данные пакетами через OTLP
		sdktrace.WithResource(res), // Передаём ресурс с атрибутами
	)

	// Устанавливаем глобальный провайдер трейсинга
	otel.SetTracerProvider(TracerProvider)
	return nil
}

// ShutdownTracing завершает работу провайдера
func ShutdownTracing(ctx context.Context) {
	if err := TracerProvider.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down TracerProvider: %v", err)
	}
}
