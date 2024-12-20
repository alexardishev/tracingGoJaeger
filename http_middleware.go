package tracing_lib

import (
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware создает мидлварь для трассировки HTTP-запросов.
func HTTPMiddleware(next http.Handler, tracerName string) http.Handler {
	// Получаем трассер с заданным именем
	tracer := otel.Tracer(tracerName)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем контекст родительского трейсинга (если он есть)
		ctx := r.Context()
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(r.Header))

		// Создаем новый спан для обработки HTTP-запроса
		ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path,
			trace.WithAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.client_ip", r.RemoteAddr),
			),
		)
		defer span.End()

		// Обновляем контекст запроса
		r = r.WithContext(ctx)

		// Оборачиваем ResponseWriter для записи статуса ответа
		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		// Вызываем следующий обработчик
		next.ServeHTTP(rw, r)

		span.SetAttributes(attribute.Int("http.status_code", rw.statusCode))
		if rw.statusCode >= 500 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", rw.statusCode))
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
	})
}

// responseWriterWrapper оборачивает http.ResponseWriter, чтобы отслеживать статус ответа
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}
