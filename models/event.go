package models

import (
	"fmt"
	"time"

	"github.com/rm-hull/street-manager-relay/generated"
)

type Event struct {
	ID              int64  `json:"-"`
	ObjectReference string `json:"-"`
	EventType       string `json:"event_type"`

	// Core location and authority info
	USRN                    *string `json:"usrn"`
	StreetName              *string `json:"street_name"`
	AreaName                *string `json:"area_name"`
	Town                    *string `json:"town"`
	HighwayAuthority        *string `json:"highway_authority"`
	HighwayAuthoritySWACode *string `json:"highway_authority_swa_code"`

	// Activity / work / permit references
	ActivityReferenceNumber  *string `json:"activity_reference_number"`
	WorkReferenceNumber      *string `json:"work_reference_number"`
	Section58ReferenceNumber *string `json:"section_58_reference_number"`
	PermitReferenceNumber    *string `json:"permit_reference_number"`
	PromoterSWACode          *string `json:"promoter_swa_code"`
	PromoterOrganisation     *string `json:"promoter_organisation"`

	// Coordinates & descriptions
	ActivityCoordinates         *string `json:"activity_coordinates"`
	ActivityLocationType        *string `json:"activity_location_type"`
	ActivityLocationDescription *string `json:"activity_location_description"`
	WorksLocationCoordinates    *string `json:"works_location_coordinates"`
	WorksLocationType           *string `json:"works_location_type"`
	Section58Coordinates        *string `json:"section_58_coordinates"`
	Section58LocationType       *string `json:"section_58_location_type"`

	// Categories & types
	WorkCategory                    *string `json:"work_category"`
	WorkCategoryRef                 *string `json:"work_category_ref"`
	WorkStatus                      *string `json:"work_status"`
	WorkStatusRef                   *string `json:"work_status_ref"`
	TrafficManagementType           *string `json:"traffic_management_type"`
	TrafficManagementTypeRef        *string `json:"traffic_management_type_ref"`
	CurrentTrafficManagementType    *string `json:"current_traffic_management_type"`
	CurrentTrafficManagementTypeRef *string `json:"current_traffic_management_type_ref"`
	RoadCategory                    *string `json:"road_category"`
	ActivityType                    *string `json:"activity_type"`
	ActivityTypeDetails             *string `json:"activity_type_details"`
	Section58Status                 *string `json:"section_58_status"`
	Section58Duration               *string `json:"section_58_duration"`
	Section58Extent                 *string `json:"section_58_extent"`

	// Dates/times
	ProposedStartDate                  *time.Time `json:"proposed_start_date"`
	ProposedEndDate                    *time.Time `json:"proposed_end_date"`
	ProposedStartTime                  *time.Time `json:"proposed_start_time"`
	ProposedEndTime                    *time.Time `json:"proposed_end_time"`
	ActualStartDateTime                *time.Time `json:"actual_start_date_time"`
	ActualEndDateTime                  *time.Time `json:"actual_end_date_time"`
	StartDate                          *time.Time `json:"start_date"`
	StartTime                          *time.Time `json:"start_time"`
	EndDate                            *time.Time `json:"end_date"`
	EndTime                            *time.Time `json:"end_time"`
	CurrentTrafficManagementUpdateDate *time.Time `json:"current_traffic_management_update_date"`

	// Flags / booleans stored as text
	IsTTRORequired            *string `json:"is_ttro_required"`
	IsCovid19Response         *string `json:"is_covid_19_response"`
	IsTrafficSensitive        *string `json:"is_traffic_sensitive"`
	IsDeemed                  *string `json:"is_deemed"`
	CollaborativeWorking      *string `json:"collaborative_working"`
	Cancelled                 *string `json:"cancelled"`
	TrafficManagementRequired *string `json:"traffic_management_required"`

	// Misc attributes
	PermitConditions     *string `json:"permit_conditions"`
	PermitStatus         *string `json:"permit_status"`
	CollaborationType    *string `json:"collaboration_type"`
	CollaborationTypeRef *string `json:"collaboration_type_ref"`
	CloseFootway         *string `json:"close_footway"`
	CloseFootwayRef      *string `json:"close_footway_ref"`
}

func (event *Event) BoundingBox() (*BBox, error) {
	if event.ActivityCoordinates == nil || *event.ActivityCoordinates == "" {
		return nil, fmt.Errorf("activity coordinates are required to calculate bounding box")
	}

	return BoundingBoxFromWKT(*event.ActivityCoordinates)
}

func NewActivityFrom(event generated.EventNotifierMessage) *Event {
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
		ActivityReferenceNumber: objectData.ActivityReferenceNumber,

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
