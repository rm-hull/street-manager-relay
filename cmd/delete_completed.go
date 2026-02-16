package cmd

import (
	"fmt"
	"log"

	"github.com/rm-hull/street-manager-relay/internal"
)

func DeleteCompletedEvents(dbPath string, days int, dryRun bool) error {
	log.Printf("Starting deletion of completed events older than %d days from %s", days, dbPath)

	repo, err := internal.NewDbRepository(dbPath)
	if err != nil {
		return fmt.Errorf("failed to create DbRepository: %w", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing DbRepository: %v", err)
		}
	}()

	ids, err := repo.GetCompletedEvents(days)
	if err != nil {
		return fmt.Errorf("failed to count completed events: %w", err)
	}

	if dryRun {
		log.Printf("[DRY RUN] Would delete %d completed events and associated rtree entries.", len(ids))
		if len(ids) > 10 {
			log.Printf("[DRY RUN] Event IDs (first 10): %v", ids[:10])
		} else if len(ids) > 0 {
			log.Printf("[DRY RUN] Event IDs: %v", ids)
		}
		return nil
	}

	log.Printf("Found %d completed events older than %d days to delete.", len(ids), days)
	count, err := repo.DeleteEvents(ids)
	if err != nil {
		return fmt.Errorf("failed to delete completed events: %w", err)
	}

	log.Printf("Successfully deleted %d completed events and associated rtree entries.", count)
	return nil
}
