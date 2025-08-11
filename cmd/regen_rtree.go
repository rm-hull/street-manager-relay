package cmd

import (
	"fmt"
	"log"

	"github.com/rm-hull/street-manager-relay/internal"
)

func RegenerateIndex(dbPath string) error {
	repo, err := internal.NewDbRepository(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize db repository: %w", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	affected, total, err := repo.RegenerateIndex(nil)
	if err != nil {
		return fmt.Errorf("error regenerating index: %w", err)
	}

	if total > 0 {
		log.Printf("Affected records: %d/%d (%.1f %%)", affected, total, float64(affected)/float64(total)*100.0)
	} else {
		log.Printf("No records found to process.")
	}

	return nil
}
