package models

type Facets struct {
	PermitStatus             []string
	TrafficManagementTypeRef []string
	WorkStatusRef            []string
	WorkCategoryRef          []string
	RoadCategory             []string
	HighwayAuthority         []string
	PromoterOrganisation     []string
}

type TemporalFilters struct {
	MaxDaysAhead  int
	MaxDaysBehind int
}

type RefData map[string]map[string]int