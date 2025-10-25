package models

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rm-hull/street-manager-relay/generated"
)

type Event struct {
	ID              int64  `json:"-"`
	ObjectReference string `json:"-"`
	EventType       string `json:"event_type"`

	// Core location and authority info
	USRN                    *string `json:"usrn,omitempty"`
	StreetName              *string `json:"street_name,omitempty"`
	AreaName                *string `json:"area_name,omitempty"`
	Town                    *string `json:"town,omitempty"`
	HighwayAuthority        *string `json:"highway_authority,omitempty"`
	HighwayAuthoritySWACode *string `json:"highway_authority_swa_code,omitempty"`

	// Activity / work / permit references
	ActivityReferenceNumber  *string `json:"activity_reference_number,omitempty"`
	WorkReferenceNumber      *string `json:"work_reference_number,omitempty"`
	Section58ReferenceNumber *string `json:"section_58_reference_number,omitempty"`
	PermitReferenceNumber    *string `json:"permit_reference_number,omitempty"`
	PromoterSWACode          *string `json:"promoter_swa_code,omitempty"`
	PromoterOrganisation     *string `json:"promoter_organisation,omitempty"`

	// Coordinates & descriptions
	ActivityCoordinates         *string `json:"activity_coordinates,omitempty"`
	ActivityLocationType        *string `json:"activity_location_type,omitempty"`
	ActivityLocationDescription *string `json:"activity_location_description,omitempty"`
	WorksLocationCoordinates    *string `json:"works_location_coordinates,omitempty"`
	WorksLocationType           *string `json:"works_location_type,omitempty"`
	Section58Coordinates        *string `json:"section_58_coordinates,omitempty"`
	Section58LocationType       *string `json:"section_58_location_type,omitempty"`

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
	Section58Status                 *string `json:"section_58_status,omitempty"`
	Section58Duration               *string `json:"section_58_duration,omitempty"`
	Section58Extent                 *string `json:"section_58_extent,omitempty"`

	// Dates/times
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

	// Flags / booleans stored as text
	IsTTRORequired            *string `json:"is_ttro_required,omitempty"`
	IsCovid19Response         *string `json:"is_covid_19_response,omitempty"`
	IsTrafficSensitive        *string `json:"is_traffic_sensitive,omitempty"`
	IsDeemed                  *string `json:"is_deemed,omitempty"`
	CollaborativeWorking      *string `json:"collaborative_working,omitempty"`
	Cancelled                 *string `json:"cancelled,omitempty"`
	TrafficManagementRequired *string `json:"traffic_management_required,omitempty"`

	// Misc attributes
	PermitConditions     *string `json:"permit_conditions,omitempty"`
	PermitStatus         *string `json:"permit_status,omitempty"`
	CollaborationType    *string `json:"collaboration_type,omitempty"`
	CollaborationTypeRef *string `json:"collaboration_type_ref,omitempty"`
	CloseFootway         *string `json:"close_footway,omitempty"`
	CloseFootwayRef      *string `json:"close_footway_ref,omitempty"`
}

func (event *Event) BoundingBox() (*BBox, error) {
	fields := []*string{
		event.WorksLocationCoordinates,
		event.ActivityCoordinates,
		event.Section58Coordinates,
	}

	for _, coords := range fields {
		if coords != nil && *coords != "" {
			return BoundingBoxFromWKT(*coords)
		}
	}

	return nil, errors.New("no coordinates found for bounding box calculation")
}

func NewEventFrom(event generated.EventNotifierMessage) *Event {
	objectData := event.ObjectData
	// Convert the generated EventNotifierMessage to our event model
	return &Event{
		ObjectReference: event.ObjectReference,
		EventType:       string(event.EventType),

		// Core location and authority info
		USRN:                    &objectData.Usrn,
		StreetName:              &objectData.StreetName,
		AreaName:                &objectData.AreaName,
		Town:                    objectData.Town,
		HighwayAuthority:        &objectData.HighwayAuthority,
		HighwayAuthoritySWACode: &objectData.HighwayAuthoritySwaCode,

		// Activity / work / permit references
		ActivityReferenceNumber:  objectData.ActivityReferenceNumber,
		PermitReferenceNumber:    objectData.PermitReferenceNumber,
		WorkReferenceNumber:      objectData.WorkReferenceNumber,
		Section58ReferenceNumber: objectData.Section58_ReferenceNumber,
		PromoterSWACode:          objectData.PromoterSwaCode,
		PromoterOrganisation:     objectData.PromoterOrganisation,

		// Coordinates & descriptions
		WorksLocationCoordinates:    objectData.WorksLocationCoordinates,
		ActivityCoordinates:         objectData.ActivityCoordinates,
		WorksLocationType:           objectData.WorksLocationType,
		ActivityLocationType:        objectData.ActivityLocationType,
		ActivityLocationDescription: objectData.ActivityLocationDescription,
		Section58Coordinates:        objectData.Section58_Coordinates,
		Section58LocationType:       objectData.Section58_LocationType,

		// Categories & types
		WorkCategory:                    objectData.WorkCategory,
		WorkCategoryRef:                 (*string)(objectData.WorkCategoryRef),
		WorkStatus:                      (*string)(objectData.WorkStatus),
		WorkStatusRef:                   (*string)(objectData.WorkStatusRef),
		TrafficManagementType:           objectData.TrafficManagementType,
		TrafficManagementTypeRef:        (*string)(objectData.TrafficManagementTypeRef),
		CurrentTrafficManagementType:    (*string)(objectData.CurrentTrafficManagementType),
		CurrentTrafficManagementTypeRef: (*string)(objectData.CurrentTrafficManagementTypeRef),
		RoadCategory:                    (*string)(objectData.RoadCategory),
		ActivityType:                    objectData.ActivityType,
		ActivityTypeDetails:             objectData.ActivityTypeDetails,
		Section58Status:                 (*string)(objectData.Section58_Status),
		Section58Duration:               (*string)(objectData.Section58_Duration),
		Section58Extent:                 (*string)(objectData.Section58_Extent),

		// Dates/times
		ProposedStartDate:                  objectData.ProposedStartDate,
		ProposedEndDate:                    objectData.ProposedEndDate,
		ProposedStartTime:                  objectData.ProposedStartTime,
		ProposedEndTime:                    objectData.ProposedEndTime,
		ActualStartDateTime:                objectData.ActualStartDateTime,
		ActualEndDateTime:                  objectData.ActualEndDateTime,
		StartDate:                          objectData.StartDate,
		StartTime:                          objectData.StartTime,
		EndDate:                            objectData.EndDate,
		EndTime:                            objectData.EndTime,
		CurrentTrafficManagementUpdateDate: objectData.CurrentTrafficManagementUpdateDate,

		// Flags / booleans stored as text
		IsTTRORequired:            (*string)(objectData.IsTtroRequired),
		IsCovid19Response:         (*string)(objectData.IsCovid19_Response),
		IsTrafficSensitive:        objectData.IsTrafficSensitive,
		IsDeemed:                  (*string)(objectData.IsDeemed),
		CollaborativeWorking:      (*string)(objectData.CollaborativeWorking),
		Cancelled:                 (*string)(objectData.Cancelled),
		TrafficManagementRequired: objectData.TrafficManagementRequired,

		// Misc attributes
		PermitConditions:     objectData.PermitConditions,
		PermitStatus:         (*string)(objectData.PermitStatus),
		CollaborationType:    (*string)(objectData.CollaborationType),
		CollaborationTypeRef: (*string)(objectData.CollaborationTypeRef),
		CloseFootway:         (*string)(objectData.CloseFootway),
		CloseFootwayRef:      (*string)(objectData.CloseFootwayRef),
	}
}
