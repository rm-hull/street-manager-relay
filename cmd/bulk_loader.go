package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rm-hull/street-manager-relay/generated"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
	"github.com/schollz/progressbar/v3"
)

func isRunningInDocker() bool {
	_, err := os.Stat("/.dockerenv")
	return !os.IsNotExist(err)
}

func BulkLoader(dbPath string, folder string, maxRecords int) error {
	repo, err := internal.NewDbRepository(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize db repository: %w", err)
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	files, err := walkFiles(folder, maxRecords)
	if err != nil {
		return fmt.Errorf("failed to import data: %w", err)
	}

	isDocker := isRunningInDocker()
	totalRecords := int64(len(files))
	var bar *progressbar.ProgressBar
	if isDocker {
		log.Println("Detected likely running inside docker container")
		bar = progressbar.DefaultSilent(totalRecords)
	} else {
		bar = progressbar.Default(totalRecords)
	}

	batch, err := repo.BatchUpsert()
	if err != nil {
		return fmt.Errorf("failed to create batch upserter: %w", err)
	}
	for idx, file := range files {
		if err := bar.Add(1); err != nil {
			return fmt.Errorf("issue with progress bar: %w", batch.Abort(err))
		}
		event, err := loadJson(file)
		if err != nil {
			return fmt.Errorf("could not load file %s: %w", file, batch.Abort(err))
		}

		_, err = batch.Upsert(models.NewActivityFrom(*event))
		if err != nil {
			return fmt.Errorf("failed to upsert activity from file %s: %w", file, batch.Abort(err))
		}

		if isDocker && idx%37 == 0 {
			log.Printf("Processed %d records...\n", idx)
		}
	}
	return batch.Done()
}

// walkFiles recursively walks through a folder and returns the relative paths for files.
func walkFiles(root string, maxFiles int) ([]string, error) {
	var files []string

	// Walk through the root directory and subdirectories.
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only add files, not directories.
		if !info.IsDir() && len(files) < maxFiles {
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func loadJson(filename string) (*generated.EventNotifierMessage, error) {

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %v", err)
	}

	event, err := generated.UnmarshalEventNotifierMessage(fileContent)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal JSON: %v", err)
	}

	return &event, nil
}
