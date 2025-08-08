package models

import (
	"fmt"
	"time"

	"github.com/rm-hull/street-manager-relay/generated"
)

type Activity struct {
	ID                      int64   `json:"-"`
	ActivityReferenceNumber *string `json:"activity_reference_number,omitempty"`

	// Location & authority info
	USRN                    *string `json:"usrn,omitempty"`
	StreetName              *string `json:"street_name,omitempty"`
	AreaName                *string `json:"area_name,omitempty"`
	Town                    *string `json:"town,omitempty"`
	HighwayAuthority        *string `json:"highway_authority,omitempty"`
	HighwayAuthoritySwaCode *string `json:"highway_authority_swa_code,omitempty"`

	// Coordinates & descriptions
	ActivityCoordinates         *string `json:"activity_coordinates,omitempty"`
	ActivityLocationType        *string `json:"activity_location_type,omitempty"`
	ActivityLocationDescription *string `json:"activity_location_description,omitempty"`

	// Categories & types
	WorkCategory                    *string `json:"work_category,omitempty"`
	WorkCategoryRef                 *string `json:"work_category_ref,omitempty"`
	WorkStatus                      *string `json:"work_status,omitempty"`
	WorkStatusRef                   *string `json:"work_status_ref,omitempty"`
	TrafficManagementType           *string `json:"traffic_management_type,omitempty"`
	TrafficManagementTypeRef        *string `json:"traffic_management_type_ref,omitempty"`
	CurrentTrafficManagementType    *string `json:"current_traffic_management_type,omitempty"`
	CurrentTrafficManagementTypeRef *string `json:"current_traffic_management_type_ref,omitempty"`
	RoadCategory                    *string `json:"road_category,omitempty"`
	ActivityType                    *string `json:"activity_type,omitempty"`
	ActivityTypeDetails             *string `json:"activity_type_details,omitempty"`

	// Dates/times (stored as ISO8601 in DB). Use pointers to allow NULL.
	ProposedStartDate                  *time.Time `json:"proposed_start_date,omitempty"`
	ProposedEndDate                    *time.Time `json:"proposed_end_date,omitempty"`
	ProposedStartTime                  *time.Time `json:"proposed_start_time,omitempty"`
	ProposedEndTime                    *time.Time `json:"proposed_end_time,omitempty"`
	ActualStartDateTime                *time.Time `json:"actual_start_date_time,omitempty"`
	ActualEndDateTime                  *time.Time `json:"actual_end_date_time,omitempty"`
	StartDate                          *time.Time `json:"start_date,omitempty"`
	StartTime                          *time.Time `json:"start_time,omitempty"`
	EndDate                            *time.Time `json:"end_date,omitempty"`
	EndTime                            *time.Time `json:"end_time,omitempty"`
	CurrentTrafficManagementUpdateDate *time.Time `json:"current_traffic_management_update_date,omitempty"`

	// Flags / booleans stored as text in DB (e.g., "Yes"/"No"/"Not provided")
	IsTtroRequired            *string `json:"is_ttro_required,omitempty"`
	IsCovid19Response         *string `json:"is_covid_19_response,omitempty"`
	IsTrafficSensitive        *string `json:"is_traffic_sensitive,omitempty"`
	IsDeemed                  *string `json:"is_deemed,omitempty"`
	CollaborativeWorking      *string `json:"collaborative_working,omitempty"`
	Cancelled                 *string `json:"cancelled,omitempty"`
	TrafficManagementRequired *string `json:"traffic_management_required,omitempty"`

	// Misc attributes
	WorksLocationCoordinates *string `json:"works_location_coordinates,omitempty"`
	WorksLocationType        *string `json:"works_location_type,omitempty"`
	PermitConditions         *string `json:"permit_conditions,omitempty"`
	CollaborationType        *string `json:"collaboration_type,omitempty"`
	CollaborationTypeRef     *string `json:"collaboration_type_ref,omitempty"`
	CloseFootway             *string `json:"close_footway,omitempty"`
	CloseFootwayRef          *string `json:"close_footway_ref,omitempty"`
}

func (a *Activity) BoundingBox() (*BBox, error) {
	if a.ActivityCoordinates == nil || *a.ActivityCoordinates == "" {
		return nil, fmt.Errorf("activity coordinates are required to calculate bounding box")
	}

	return BoundingBoxFromWKT(*a.ActivityCoordinates)
}

func NewActivityFrom(event generated.EventNotifierMessage) *Activity {
	// Convert the generated EventNotifierMessage to our Activity model
	return &Activity{
		ActivityReferenceNumber:            event.ObjectData.ActivityReferenceNumber,
		USRN:                               &event.ObjectData.Usrn,
		StreetName:                         &event.ObjectData.StreetName,
		AreaName:                           &event.ObjectData.AreaName,
		Town:                               event.ObjectData.Town,
		HighwayAuthority:                   &event.ObjectData.HighwayAuthority,
		HighwayAuthoritySwaCode:            &event.ObjectData.HighwayAuthoritySwaCode,
		ActivityCoordinates:                event.ObjectData.ActivityCoordinates,
		ActivityLocationType:               event.ObjectData.ActivityLocationType,
		ActivityLocationDescription:        event.ObjectData.ActivityLocationDescription,
		WorkCategory:                       event.ObjectData.WorkCategory,
		WorkCategoryRef:                    ToString(event.ObjectData.WorkCategoryRef),
		WorkStatus:                         ToString(event.ObjectData.WorkStatus),
		WorkStatusRef:                      ToString(event.ObjectData.WorkStatusRef),
		TrafficManagementType:              event.ObjectData.TrafficManagementType,
		TrafficManagementTypeRef:           ToString(event.ObjectData.TrafficManagementTypeRef),
		CurrentTrafficManagementType:       ToString(event.ObjectData.CurrentTrafficManagementType),
		CurrentTrafficManagementTypeRef:    ToString(event.ObjectData.CurrentTrafficManagementTypeRef),
		RoadCategory:                       ToString(event.ObjectData.RoadCategory),
		ActivityType:                       event.ObjectData.ActivityType,
		ActivityTypeDetails:                event.ObjectData.ActivityTypeDetails,
		ProposedStartDate:                  event.ObjectData.ProposedStartDate,
		ProposedEndDate:                    event.ObjectData.ProposedEndDate,
		ProposedStartTime:                  event.ObjectData.ProposedStartTime,
		ProposedEndTime:                    event.ObjectData.ProposedEndTime,
		ActualStartDateTime:                event.ObjectData.ActualStartDateTime,
		ActualEndDateTime:                  event.ObjectData.ActualEndDateTime,
		StartDate:                          event.ObjectData.StartDate,
		StartTime:                          event.ObjectData.StartTime,
		EndDate:                            event.ObjectData.EndDate,
		EndTime:                            event.ObjectData.EndTime,
		CurrentTrafficManagementUpdateDate: event.ObjectData.CurrentTrafficManagementUpdateDate,
		IsTtroRequired:                     ToString(event.ObjectData.IsTtroRequired),
		IsCovid19Response:                  ToString(event.ObjectData.IsCovid19_Response),
		IsTrafficSensitive:                 event.ObjectData.IsTrafficSensitive,
		IsDeemed:                           ToString(event.ObjectData.IsDeemed),
		CollaborativeWorking:               ToString(event.ObjectData.CollaborativeWorking),
		Cancelled:                          ToString(event.ObjectData.Cancelled),
		TrafficManagementRequired:          event.ObjectData.TrafficManagementRequired,
		WorksLocationCoordinates:           event.ObjectData.WorksLocationCoordinates,
		WorksLocationType:                  event.ObjectData.WorksLocationType,
		PermitConditions:                   event.ObjectData.PermitConditions,
		CollaborationType:                  ToString(event.ObjectData.CollaborationType),
		CollaborationTypeRef:               ToString(event.ObjectData.CollaborationTypeRef),
		CloseFootway:                       ToString(event.ObjectData.CloseFootway),
		CloseFootwayRef:                    ToString(event.ObjectData.CloseFootwayRef),
	}
}

type StringType interface {
	~string
}

func ToString[T StringType](ptr *T) *string {
	if ptr == nil {
		return nil
	}
	str := string(*ptr)
	return &str
}
