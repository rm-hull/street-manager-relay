CREATE TABLE IF NOT EXISTS activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- Unique reference from source data
    activity_reference_number TEXT UNIQUE,

    -- Location & authority info
    usrn TEXT,
    street_name TEXT,
    area_name TEXT,
    town TEXT,
    highway_authority TEXT,
    highway_authority_swa_code TEXT,

    -- Coordinates & descriptions
    activity_coordinates TEXT,
    activity_location_type TEXT,
    activity_location_description TEXT,

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

    -- Dates/times
    proposed_start_date TEXT,
    proposed_end_date TEXT,
    proposed_start_time TEXT,
    proposed_end_time TEXT,
    actual_start_date_time TEXT,
    actual_end_date_time TEXT,
    start_date TEXT,
    start_time TEXT,
    end_date TEXT,
    end_time TEXT,
    current_traffic_management_update_date TEXT,

    -- Flags / booleans stored as text
    is_ttro_required TEXT,
    is_covid_19_response TEXT,
    is_traffic_sensitive TEXT,
    is_deemed TEXT,
    collaborative_working TEXT,
    cancelled TEXT,
    traffic_management_required TEXT,

    -- Misc attributes
    works_location_coordinates TEXT,
    works_location_type TEXT,
    permit_conditions TEXT,
    collaboration_type TEXT,
    collaboration_type_ref TEXT,
    close_footway TEXT,
    close_footway_ref TEXT
);


-- Composite index for actual date range queries
CREATE INDEX IF NOT EXISTS idx_events_actual_range
    ON activities(actual_start_date_time, actual_end_date_time);

-- Composite index for planned date range queries
CREATE INDEX IF NOT EXISTS idx_events_planned_range
    ON activities(start_date, end_date);

-- Unique index (implicit from UNIQUE constraint, but can be named explicitly)
CREATE UNIQUE INDEX IF NOT EXISTS idx_activity_ref ON activities(activity_reference_number);


-- R-Tree index table for bounding boxes
CREATE VIRTUAL TABLE IF NOT EXISTS activities_rtree USING rtree(
    id,    -- matches activities.id
    minx,
    maxx,
    miny,
    maxy
);