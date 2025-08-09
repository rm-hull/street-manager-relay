CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    object_reference TEXT UNIQUE,
    event_type TEXT,

    -- Core location and authority info
    usrn TEXT,
    street_name TEXT,
    area_name TEXT,
    town TEXT,
    highway_authority TEXT,
    highway_authority_swa_code TEXT,

    -- Activity / work / permit references
    activity_reference_number TEXT,
    work_reference_number TEXT,
    permit_reference_number TEXT,
    promoter_swa_code TEXT,
    promoter_organisation TEXT,

    -- Coordinates & descriptions
    works_location_coordinates TEXT,
    activity_coordinates TEXT,
    works_location_type TEXT,
    activity_location_type TEXT,
    activity_location_description TEXT,
    section_58_coordinates TEXT,

    -- Categories & types
    work_category TEXT,
    work_category_ref TEXT,
    work_status TEXT,
    work_status_ref TEXT,
    traffic_management_type TEXT,
    traffic_management_type_ref TEXT,
    current_traffic_management_type TEXT,
    current_traffic_management_type_ref TEXT,
    road_category TEXT,
    activity_type TEXT,
    activity_type_details TEXT,
    section_58_status TEXT,
    section_58_duration TEXT,
    section_58_extent TEXT,
    section_58_location_type TEXT,

    -- Dates/times
    proposed_start_date TIMESTAMP,
    proposed_end_date TIMESTAMP,
    proposed_start_time TIMESTAMP,
    proposed_end_time TIMESTAMP,
    actual_start_date_time TIMESTAMP,
    actual_end_date_time TIMESTAMP,
    start_date TIMESTAMP,
    start_time TIMESTAMP,
    end_date TIMESTAMP,
    end_time TIMESTAMP,
    current_traffic_management_update_date TIMESTAMP,

    -- Flags / booleans stored as text
    is_ttro_required TEXT,
    is_covid_19_response TEXT,
    is_traffic_sensitive TEXT,
    is_deemed TEXT,
    collaborative_working TEXT,
    cancelled TEXT,
    traffic_management_required TEXT,

    -- Misc attributes
    permit_conditions TEXT,
    permit_status TEXT,
    collaboration_type TEXT,
    collaboration_type_ref TEXT,
    close_footway TEXT,
    close_footway_ref TEXT,
    section_58_reference_number TEXT
);

-- Composite index for actual date range queries
CREATE INDEX IF NOT EXISTS idx_events_actual_range
    ON activities(actual_start_date_time, actual_end_date_time);

-- Composite index for planned date range queries
CREATE INDEX IF NOT EXISTS idx_events_planned_range
    ON activities(start_date, end_date);

-- Unique index (implicit from UNIQUE constraint, but can be named explicitly)
CREATE UNIQUE INDEX IF NOT EXISTS idx_events_ref ON events(ref);

-- R-Tree index table for bounding boxes
CREATE VIRTUAL TABLE IF NOT EXISTS events_rtree USING rtree(
    id,    -- matches events.id
    minx,
    maxx,
    miny,
    maxy
);