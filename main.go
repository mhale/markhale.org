package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const port int = 8080        // External port for website; fly.toml forwards ports 80 and 443 to 8080.
const metricsPort int = 9091 // Internal port for Prometheus exporter.

//go:embed index.html
var indexHTML string

//go:embed style.css
var styleCSS string

//go:embed mta-sts.txt
var mtaSTSPolicy string

var httpReqs = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "http_requests_total",
	Help: "HTTP requests processed, partitioned by protocol and hostname.",
}, []string{"proto", "host"})

func main() {
	// Tweak the logs if deployed.
	_, onFly := os.LookupEnv("FLY_APP_NAME")
	if onFly {
		log.SetFlags(0)    // Remove duplicated timestamp
		log.SetPrefix(" ") // Differentiate between site and Docker entries
	}

	log.Println("Starting")
	defer log.Println("Exiting")

	// Configure Prometheus exporter.
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())

	metricsSrv := &http.Server{
		Addr:         fmt.Sprintf(":%d", metricsPort),
		Handler:      metricsMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go func() { log.Fatal(metricsSrv.ListenAndServe()) }()

	// Configure website.
	mux := &http.ServeMux{}
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/style.css", styleHandler)
	mux.HandleFunc("/.well-known/mta-sts.txt", mtaSTSHandler)

	// Configure middleware.
	var handler http.Handler = mux
	handler = filter(redirect(serve(handler)))
	// Suppress log entry "http: URL query contains semicolon..." emitted in server.go.
	// It converts ';' in the query string to '&'.
	handler = http.AllowQuerySemicolons(handler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  120 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
	}

	idleConnsClosed := make(chan struct{})

	// Graceful shutdown handler.
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		log.Println("Received SIGINT")
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("Shutdown error: %v", err)
		}

		close(idleConnsClosed)
	}()

	log.Printf("Listening on :%d", port)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("ListenAndServe error: %v", err)
	}

	<-idleConnsClosed
}
