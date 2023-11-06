package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const name = "foo"

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider() (func(context.Context) error, error) {
	ctx := context.Background()
	traceExporter, err := otlptracegrpc.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer(name).Start(r.Context(), "root")
	defer span.End()
	doSomething(ctx)
	w.Write([]byte("Foo is up!"))
}

func getError(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer(name).Start(r.Context(), "error")
	defer span.End()
	doAnotherThing(ctx)
	w.Write([]byte("Foo is up!"))
}

func getLong(w http.ResponseWriter, r *http.Request) {
	ctx, span := otel.Tracer(name).Start(r.Context(), "long")
	defer span.End()
	doLongThing(ctx)
	w.Write([]byte("Foo is up!"))
}

func doSomething(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "do something")
	defer span.End()
}

func doAnotherThing(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "do another thing")
	span.SetStatus(codes.Error, "do another thing failed")
	defer span.End()
}

func doLongThing(ctx context.Context) {
	_, span := otel.Tracer(name).Start(ctx, "do long thing")
	defer span.End()

	time.Sleep(1000 * time.Millisecond)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	shutdown, err := initProvider()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatal("failed to shutdown TracerProvider: %w", err)
		}
	}()

	rootHandler := otelhttp.NewHandler(http.HandlerFunc(getRoot), "Hello")
	errorHandler := otelhttp.NewHandler(http.HandlerFunc(getError), "Error")
	longHandler := otelhttp.NewHandler(http.HandlerFunc(getLong), "Long")

	http.Handle("/", rootHandler)
	http.Handle("/error", errorHandler)
	http.Handle("/long", longHandler)

	err = http.ListenAndServe(":8081", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
