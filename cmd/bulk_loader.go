package cmd

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
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
		return errors.Wrap(err, "failed to initialize db repository")
	}
	defer func() {
		if err := repo.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	log.Println("Finding files to import...")
	files, err := walkFiles(folder, maxRecords)
	if err != nil {
		return errors.Wrap(err, "failed to import data")
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
		return errors.Wrap(err, "failed to create batch upserter")
	}
	for idx, file := range files {
		if err := bar.Add(1); err != nil {
			return errors.Wrap(batch.Abort(err), "issue with progress bar")
		}
		event, err := loadJson(file)
		if err != nil {
			return errors.Wrapf(batch.Abort(err), "could not load file %s", file)
		}

		_, err = batch.Upsert(models.NewEventFrom(*event))
		if err != nil {
			return errors.Wrapf(batch.Abort(err), "failed to upsert event from file %s", file)
		}

		if isDocker && idx%37 == 0 {
			log.Printf("Processed %d records...\n", idx)
		}
	}
	return batch.Done()
}

// walkFiles recursively walks through a folder and returns the relative paths for files.
func walkFiles(root string, maxFiles int) ([]string, error) {
	files := make([]string, 0, 1000)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if len(files) >= maxFiles {
			return fs.SkipAll
		}

		if !info.IsDir() {
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
		return nil, errors.Wrap(err, "could not read file")
	}

	event, err := generated.UnmarshalEventNotifierMessage(fileContent)
	if err != nil {
		return nil, errors.Wrap(err, "could not unmarshal JSON")
	}

	return &event, nil
}
