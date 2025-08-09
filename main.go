package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/rm-hull/street-manager-relay/cmd"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/spf13/cobra"
)

func main() {
	var err error
	var dbPath string
	var port int
	var debug bool

	internal.ShowVersion()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	internal.EnvironmentVars()

	rootCmd := &cobra.Command{
		Use:  "street-manager-relay",
		Long: `Street manager relay API & data importers`,
	}

	apiServerCmd := &cobra.Command{
		Use:   "api-server [--db <path>] [--port <port>] [--debug]",
		Short: "Start HTTP API server",
		Run: func(_ *cobra.Command, _ []string) {
			cmd.ApiServer(dbPath, port, debug)
		},
	}

	apiServerCmd.Flags().StringVar(&dbPath, "db", "./data/street-manager.db", "Path to street-manager SQLite database")
	apiServerCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")
	apiServerCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging (pprof) - WARING: do not enable in production")

	rootCmd.AddCommand(apiServerCmd)
	if err = rootCmd.Execute(); err != nil {
		panic(err)
	}
}
