SELECT 'permit_status' AS facet, permit_status AS value, COUNT(*) AS cnt
FROM events
GROUP BY permit_status
UNION ALL
SELECT 'traffic_management_type_ref' AS facet, traffic_management_type_ref AS value, COUNT(*) AS cnt
FROM events
GROUP BY traffic_management_type_ref
UNION ALL
SELECT 'work_status_ref' AS facet, work_status_ref AS value, COUNT(*) AS cnt
FROM events
GROUP BY work_status_ref
UNION ALL
SELECT 'work_category_ref' AS facet, work_category_ref AS value, COUNT(*) AS cnt
FROM events
GROUP BY work_category_ref
UNION ALL
SELECT 'road_category' AS facet, road_category AS value, COUNT(*) AS cnt
FROM events
GROUP BY road_category
UNION ALL
SELECT 'highway_authority' AS facet, highway_authority AS value, COUNT(*) AS cnt
FROM events
GROUP BY highway_authority
UNION ALL
SELECT 'promoter_organisation' AS facet, promoter_organisation AS value, COUNT(*) AS cnt
FROM events
GROUP BY promoter_organisation;