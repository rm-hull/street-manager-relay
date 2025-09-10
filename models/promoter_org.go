package models

import (
	_ "embed"
	"fmt"
	"strconv"
)

type PromoterOrg struct {
	Id      int
	Name    string
	Url     string
	Favicon *string
}

func (org *PromoterOrg) ToCSV() []string {
	row := []string{
		strconv.Itoa(org.Id),
		org.Name,
		org.Url,
		"",
	}
	if org.Favicon != nil {
		row[3] = *org.Favicon
	}

	return row
}

func (org *PromoterOrg) FromCSV(record, headers []string) (*PromoterOrg, error) {
	id, err := strconv.Atoi(record[0])
	if err != nil {
		return nil, fmt.Errorf("failed to convert id=%s: %w", record[0], err)
	}

	org.Id = id
	org.Name = record[1]
	org.Url = record[2]
	if len(record) == 4 && record[3] != "" {
		org.Favicon = &record[3]
	} else {
		org.Favicon = nil
	}

	return org, nil
}
