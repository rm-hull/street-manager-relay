package cmd

import (
	"log"

	"github.com/cockroachdb/errors"
	"github.com/rm-hull/street-manager-relay/internal"
)

func RegenerateIndex(dbPath string) error {
	repo, err := internal.NewDbRepository(dbPath)
	if err != nil {
		return errors.Wrap(err, "failed to initialize db repository")
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	affected, total, err := repo.RegenerateIndex()
	if err != nil {
		return errors.Wrap(err, "error regenerating index")
	}

	if total > 0 {
		log.Printf("Affected records: %d/%d (%.1f %%)", affected, total, float64(affected)/float64(total)*100.0)
	} else {
		log.Printf("No records found to process.")
	}

	return nil
}
