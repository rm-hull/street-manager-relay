SELECT
    -- Identifiers
    e.id,
    e.event_type,
    e.object_reference,
    e.activity_reference_number,
    e.work_reference_number,
    e.section_58_reference_number,
    e.permit_reference_number,

    -- Core location and authority info
    e.usrn,
    e.street_name,
    e.area_name,
    e.town,
    e.highway_authority,
    e.highway_authority_swa_code,
    e.promoter_swa_code,
    e.promoter_organisation,

    -- Coordinates & descriptions
    e.activity_coordinates,
    e.activity_location_type,
    e.activity_location_description,
    e.works_location_coordinates,
    e.works_location_type,
    e.section_58_coordinates,
    e.section_58_location_type,

    -- Categories & types
    e.work_category,
    e.work_category_ref,
    e.work_status,
    e.work_status_ref,
    e.traffic_management_type,
    e.traffic_management_type_ref,
    e.current_traffic_management_type,
    e.current_traffic_management_type_ref,
    e.road_category,
    e.activity_type,
    e.activity_type_details,
    e.section_58_status,
    e.section_58_duration,
    e.section_58_extent,

    -- Dates/times
    e.proposed_start_date,
    e.proposed_end_date,
    e.proposed_start_time,
    e.proposed_end_time,
    e.actual_start_date_time,
    e.actual_end_date_time,
    e.start_date,
    e.start_time,
    e.end_date,
    e.end_time,
    e.current_traffic_management_update_date,

    -- Flags / booleans stored as text
    e.is_ttro_required,
    e.is_covid_19_response,
    e.is_traffic_sensitive,
    e.is_deemed,
    e.collaborative_working,
    e.cancelled,
    e.traffic_management_required,

    -- Misc attributes
    e.permit_conditions,
    e.permit_status,
    e.collaboration_type,
    e.collaboration_type_ref,
    e.close_footway,
    e.close_footway_ref

FROM events AS e
INNER JOIN events_rtree r ON e.id = r.id
WHERE r.minx <= ? AND r.maxx >= ? AND r.miny <= ? AND r.maxy >= ?
  AND (
      -- Get the first available start and end datetime
      (
        COALESCE(e.actual_start_date_time,
                 e.start_date,
                 e.proposed_start_date) >= CURRENT_DATE
      )
      OR
      (
        COALESCE(e.actual_start_date_time,
                 e.start_date,
                 e.proposed_start_date) <= CURRENT_DATE
        AND (
          COALESCE(e.actual_end_date_time,
                   e.end_date,
                   e.proposed_end_date) IS NULL
          OR COALESCE(e.actual_end_date_time,
                      e.end_date,
                      e.proposed_end_date) >= CURRENT_DATE
        )
      )
  );
