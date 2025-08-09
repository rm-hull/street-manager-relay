package cmd

import (
	"fmt"
	"log"

	"github.com/rm-hull/street-manager-relay/internal"
)

func BulkLoader(dbPath string, folder string) error {
	repo, err := internal.NewDbRepository(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize db repository: %w", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	return nil
}
