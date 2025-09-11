package models

import (
	_ "embed"
)

type PromoterOrg struct {
	Id      string
	Name    string
	Url     string
	Favicon *string
}

func (org *PromoterOrg) ToCSV() []string {
	row := []string{
		org.Id,
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
	org.Id = record[0]
	org.Name = record[1]
	org.Url = record[2]
	if len(record) == 4 && record[3] != "" {
		org.Favicon = &record[3]
	} else {
		org.Favicon = nil
	}

	return org, nil
}
