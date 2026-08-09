package main

import (
	"log/slog"
	"math"
	"os"

	"github.com/joho/godotenv"
	"github.com/map-services/street-manager-relay/cmd"
	"github.com/spf13/cobra"
)

func main() {
	var err error
	var dbPath string
	var port int
	var debug bool
	var maxFiles int
	var days int
	var dryRun bool
	var filePath string
	var excludeActivityTypes []string

	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found")
	}

	rootCmd := &cobra.Command{
		Use:  "street-manager-relay",
		Long: `Street manager relay API & data importers`,
	}

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "./data/street-manager.db", "Path to street-manager SQLite database")

	apiServerCmd := &cobra.Command{
		Use:   "api-server [--db <path>] [--port <port>] [--debug]",
		Short: "Start HTTP API server",
		Run: func(_ *cobra.Command, _ []string) {
			cmd.ApiServer(dbPath, port, debug, excludeActivityTypes)
		},
	}

	apiServerCmd.Flags().IntVar(&port, "port", 8080, "Port to run HTTP server on")
	apiServerCmd.Flags().BoolVar(&debug, "debug", false, "Enable debugging (pprof) - WARING: do not enable in production")
	apiServerCmd.Flags().StringSliceVar(&excludeActivityTypes, "exclude-activity-types", []string{}, "Comma-separated list of activity types to exclude")

	bulkLoaderCmd := &cobra.Command{
		Use:   "bulk-loader [--db <path>] [--max-files <n>] [--exclude-activity-types <types>] <folder>",
		Short: "Run bulk data loader",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			if err := cmd.BulkLoader(dbPath, args[0], maxFiles, excludeActivityTypes); err != nil {
				slog.Error("Failed to run bulk loader", "error", err)
				os.Exit(1)
			}
		},
	}

	bulkLoaderCmd.Flags().IntVar(&maxFiles, "max-files", math.MaxInt, "Maximum number of files to process")
	bulkLoaderCmd.Flags().StringSliceVar(&excludeActivityTypes, "exclude-activity-types", []string{}, "Comma-separated list of activity types to exclude")

	regenCmd := &cobra.Command{
		Use:   "regen [--db <path>]",
		Short: "Regenerate Indexes",
		Run: func(_ *cobra.Command, _ []string) {
			if err := cmd.RegenerateIndex(dbPath); err != nil {
				slog.Error("Regenerate index failed", "error", err)
				os.Exit(1)
			}
		},
	}

	deleteCompletedCmd := &cobra.Command{
		Use:   "delete-completed [--db <path>] [--days <n>] [--dry-run]",
		Short: "Delete completed events older than N days",
		Run: func(_ *cobra.Command, _ []string) {
			if days <= 0 {
				log.Fatalf("--days must be greater than 0")
			}
			if err := cmd.DeleteCompletedEvents(dbPath, days, dryRun); err != nil {
				log.Fatalf("Failed to delete completed events: %v", err)
			}
		},
	}
	deleteCompletedCmd.Flags().IntVar(&days, "days", 30, "Delete events completed more than N days ago")
	deleteCompletedCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be deleted without making changes")

	updateFaviconsCmd := &cobra.Command{
		Use:   "favicons [--file <path>]",
		Short: "Update favicons",
		Run: func(_ *cobra.Command, _ []string) {
			if err := cmd.UpdateFaviconsInCSV(filePath); err != nil {
				slog.Error("Update favicons failed", "error", err)
				os.Exit(1)
			}
		},
	}
	updateFaviconsCmd.Flags().StringVar(&filePath, "file", "./internal/promoter/organisations.csv", "Path to promoter orgs CSV file")

	rootCmd.AddCommand(apiServerCmd)
	rootCmd.AddCommand(bulkLoaderCmd)
	rootCmd.AddCommand(regenCmd)
	rootCmd.AddCommand(deleteCompletedCmd)
	rootCmd.AddCommand(updateFaviconsCmd)
	if err = rootCmd.Execute(); err != nil {
		panic(err)
	}
}
