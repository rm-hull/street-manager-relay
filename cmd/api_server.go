package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Depado/ginprom"
	"github.com/aurowora/compress"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/kofalt/go-memoize"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/internal/promoter"
	"github.com/rm-hull/street-manager-relay/internal/routes"
	"github.com/tavsec/gin-healthcheck/checks"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"

	healthcheck "github.com/tavsec/gin-healthcheck"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

func ApiServer(dbPath string, port int, debug bool) {

	organisations, err := promoter.GetPromoterOrgsMap()
	if err != nil {
		log.Fatalf("failed to initialize promoter organisations: %v", err)
	}

	repo, err := internal.NewDbRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize db repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	err = sentry.Init(sentry.ClientOptions{
		Dsn:         os.Getenv("SENTRY_DSN"),
		Debug:       debug,
		Release:     versioninfo.Revision[:7],
		Environment: os.Getenv("MODE"),
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	defer sentry.Flush(2 * time.Second)

	r := gin.New()

	prometheus := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz"),
	)

	r.Use(
		sentrygin.New(sentrygin.Options{
			Repanic:         true,
			WaitForDelivery: false,
			Timeout:         5 * time.Second,
		}),
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz", "/metrics"),
		prometheus.Instrument(),
		compress.Compress(),
		cors.Default(),
		sentryErrorHandler(),
	)

	if debug {
		log.Println("WARNING: pprof endpoints are enabled and exposed. Do not run with this flag in production.")
		pprof.Register(r)
	}

	err = healthcheck.New(r, hc_config.DefaultConfig(), []checks.Check{
		repo.HealthCheck(),
	})
	if err != nil {
		log.Fatalf("failed to initialize healthcheck: %v", err)
	}

	certManager := internal.NewCertManager(memoize.NewMemoizer(24*time.Hour, 1*time.Hour))

	r.POST("/v1/street-manager-relay/sns", routes.HandleSNSMessage(repo, certManager))
	r.GET("/v1/street-manager-relay/search", routes.HandleSearch(repo, organisations))
	r.GET("/v1/street-manager-relay/refdata", routes.HandleRefData(repo, memoize.NewMemoizer(10*time.Minute, 1*time.Hour)))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP API Server on port %d...", port)
	err = r.Run(addr)
	log.Fatalf("HTTP API Server failed to start on port %d: %v", port, err)
}

func sentryErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			hub := sentrygin.GetHubFromContext(c)
			for _, e := range c.Errors {
				if hub != nil {
					hub.CaptureException(e.Err)
				} else {
					sentry.CaptureException(e.Err)
				}
			}
		}
	}
}
