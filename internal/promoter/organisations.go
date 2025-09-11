package promoter

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
)

//go:embed organisations.csv
var promoterOrgsCSV string

func GetPromoterOrgsList() ([]*models.PromoterOrg, error) {
	org := &models.PromoterOrg{}
	arr := make([]*models.PromoterOrg, 0, 100)
	reader := strings.NewReader(promoterOrgsCSV)

	for record := range internal.ParseCSV(reader, false, org.FromCSV) {
		if record.Error != nil {
			return nil, fmt.Errorf("failed to load promoter organisations: %w", record.Error)
		}
		copy := *record.Value
		arr = append(arr, &copy)
	}

	return arr, nil
}

func GetPromoterOrgsMap() (Organisations, error) {
	orgs, err := GetPromoterOrgsList()
	if err != nil {
		return nil, err
	}

	m := make(map[string]*models.PromoterOrg)
	for _, record := range orgs {
		if _, ok := m[record.Id]; ok {
			return nil, fmt.Errorf("duplicate key detected: %s", record.Id)
		}
		m[record.Id] = record
	}

	return m, nil
}

type Organisations map[string]*models.PromoterOrg