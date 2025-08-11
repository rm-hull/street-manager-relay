package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/Depado/ginprom"
	"github.com/aurowora/compress"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/kofalt/go-memoize"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/internal/routes"
	"github.com/tavsec/gin-healthcheck/checks"

	healthcheck "github.com/tavsec/gin-healthcheck"
	hc_config "github.com/tavsec/gin-healthcheck/config"
)

func ApiServer(dbPath string, port int, debug bool) {

	repo, err := internal.NewDbRepository(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize db repository: %v", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	r := gin.New()

	prometheus := ginprom.New(
		ginprom.Engine(r),
		ginprom.Path("/metrics"),
		ginprom.Ignore("/healthz"),
	)

	r.Use(
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/healthz", "/metrics"),
		prometheus.Instrument(),
		compress.Compress(),
		cors.Default(),
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

	cache := memoize.NewMemoizer(24*time.Hour, 1*time.Hour)
	certManager := internal.NewCertManager(cache)

	r.POST("/v1/street-manager-relay/sns", routes.HandleSNSMessage(repo, certManager))
	r.GET("/v1/street-manager-relay/search", routes.HandleSearch(repo))

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HTTP API Server on port %d...", port)
	err = r.Run(addr)
	log.Fatalf("HTTP API Server failed to start on port %d: %v", port, err)
}
