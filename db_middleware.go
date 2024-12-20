package tracing_lib

import (
	"context"

	"github.com/jackc/pgx/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// TraceDB оборачивает вызов QueryRow для логирования в трейс и возвращает pgx.Row
func TraceDB(ctx context.Context, operationName string, query string, row pgx.Row) pgx.Row {
	tracer := otel.Tracer("database")
	_, span := tracer.Start(ctx, operationName)
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operationName),
		attribute.String("db.query", query), // Добавляем сам SQL-запрос
	)

	return row
}
