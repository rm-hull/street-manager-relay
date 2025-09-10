package main

import (
	"log"
	"math"

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
	var maxFiles int
	var filePath string

	internal.ShowVersion()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	internal.EnvironmentVars()

	rootCmd := &cobra.Command{
		Use:  "street-manager-relay",
		Long: `Street manager relay API & data importers`,
	}

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./data/street-manager.db", "Path to street-manager SQLite database")

	apiServerCmd := &cobra.Command{
		Use:   "api-server [--db <path>] [--port <port>] [--debug]",
		Short: "Start HTTP API server",
		Run: func(_ *cobra.Command, _ []string) {
			cmd.ApiServer(dbPath, port, debug)
		},
	}

	apiServerCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")
	apiServerCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging (pprof) - WARING: do not enable in production")

	bulkLoaderCmd := &cobra.Command{
		Use:   "bulk-loader [--db <path>] [--max-files <n>] <folder>",
		Short: "Run bulk data loader",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			if err := cmd.BulkLoader(dbPath, args[0], maxFiles); err != nil {
				log.Fatalf("Failed to run bulk loader: %v", err)
			}
		},
	}

	bulkLoaderCmd.Flags().IntVar(&maxFiles, "max-files", math.MaxInt, "Maximum number of files to process")

	regenCmd := &cobra.Command{
		Use:   "regen [--db <path>]",
		Short: "Regenerate Indexes",
		Run: func(_ *cobra.Command, _ []string) {
			if err := cmd.RegenerateIndex(dbPath); err != nil {
				log.Fatalf("Regenerate index failed: %v", err)
			}
		},
	}

	updateFaviconsCmd := &cobra.Command{
		Use:   "favicons [--file <path>]",
		Short: "Update favicons",
		Run: func(_ *cobra.Command, _ []string) {
			if err := cmd.UpdateFaviconsInCSV(filePath); err != nil {
				log.Fatalf("Update favicons failed: %v", err)
			}
		},
	}
	updateFaviconsCmd.Flags().StringVar(&filePath, "file", "./internal/promoter/organisations.csv", "Path to promoter orgs CSV file")

	rootCmd.AddCommand(apiServerCmd)
	rootCmd.AddCommand(bulkLoaderCmd)
	rootCmd.AddCommand(regenCmd)
	rootCmd.AddCommand(updateFaviconsCmd)
	if err = rootCmd.Execute(); err != nil {
		panic(err)
	}
}
