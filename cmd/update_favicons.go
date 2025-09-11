package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/rm-hull/street-manager-relay/internal/favicon"
	"github.com/rm-hull/street-manager-relay/internal/promoter"
	"github.com/rm-hull/street-manager-relay/models"
)

func UpdateFaviconsInCSV(csvFile string) error {

	orgs, err := promoter.GetPromoterOrgsList()
	if err != nil {
		return err
	}

	updated := make([]*models.PromoterOrg, 0, len(orgs))
	for idx, record := range orgs {

		log.Printf("Processing record %d: %s", idx, record.Url)

		iconInfo, err := favicon.Extract(record.Url)
		if err != nil {
			log.Printf("failed to extract favicon for %s: %v", record.Url, err)
		} else {
			record.Favicon = &iconInfo.Href
		}
		updated = append(updated, record)
	}

	f, err := os.OpenFile(csvFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", csvFile, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	csvWriter := csv.NewWriter(f)
	defer csvWriter.Flush()

	for _, record := range updated {
		row := record.ToCSV()
		if err := csvWriter.Write(row); err != nil {
			return fmt.Errorf("failed to write row=%v: %w", row, err)
		}
	}
	return nil
}
