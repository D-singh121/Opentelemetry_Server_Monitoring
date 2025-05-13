package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

func initTracer() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	// OTLP HTTP Exporter - Send traces to Tempo
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("localhost:4318"), // Tempo endpoint
		otlptracehttp.WithInsecure(),                 // no TLS
	)
	if err != nil {
		return nil, err
	}

	// Define Resource (Service name)
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("gin-otel-raw"),
		),
	)
	if err != nil {
		return nil, err
	}

	// Create Tracer Provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp, nil
}

func helloHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("gin-otel-raw")

	// Start a span manually
	_, span := tracer.Start(ctx, "hello-handler")
	defer span.End()

	time.Sleep(50 * time.Millisecond) // Simulate processing

	span.SetAttributes(attribute.String("custom.attribute", "hello_value"))

	c.JSON(http.StatusOK, gin.H{"message": "Hello from /hello"})
}

func healthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	tracer := otel.Tracer("gin-otel-raw")

	_, span := tracer.Start(ctx, "health-handler")
	defer span.End()

	time.Sleep(30 * time.Millisecond) // Simulate processing

	span.SetAttributes(attribute.String("custom.attribute", "health_value"))

	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}

func main() {
	tp, err := initTracer()
	if err != nil {
		log.Fatalf("failed to initialize tracer: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Fatalf("error shutting down tracer provider: %v", err)
		}
	}()

	r := gin.Default()

	// Routes with manually created spans
	r.GET("/hello", helloHandler)
	r.GET("/health", healthHandler)

	fmt.Println("Server is running at http://localhost:8080")
	r.Run(":8080")
}
