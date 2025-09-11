package models

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

func FromCSV(record, headers []string) (*PromoterOrg, error) {
	org := &PromoterOrg{
		Id:   record[0],
		Name: record[1],
		Url:  record[2],
	}
	if len(record) == 4 && record[3] != "" {
		org.Favicon = &record[3]
	}
	return org, nil
}
