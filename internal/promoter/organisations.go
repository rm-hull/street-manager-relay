package promoter

import (
	_ "embed"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/rm-hull/street-manager-relay/internal"
	"github.com/rm-hull/street-manager-relay/models"
)

//go:embed organisations.csv
var promoterOrgsCSV string

func GetPromoterOrgsList() ([]*models.PromoterOrg, error) {
	arr := make([]*models.PromoterOrg, 0, 100)
	reader := strings.NewReader(promoterOrgsCSV)

	for record := range internal.ParseCSV(reader, false, models.FromCSV) {
		if record.Error != nil {
			return nil, errors.Wrap(record.Error, "failed to load promoter organisations")
		}
		arr = append(arr, record.Value)
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
			return nil, errors.Newf("duplicate key detected: %s", record.Id)
		}
		m[record.Id] = record
	}

	return m, nil
}

type Organisations map[string]*models.PromoterOrg
