SELECT 'permit_status' AS facet, permit_status AS value, COUNT(*) AS cnt
FROM events
GROUP BY permit_status
UNION ALL
SELECT 'traffic_management_type_ref', traffic_management_type_ref, COUNT(*)
FROM events
GROUP BY traffic_management_type_ref
UNION ALL
SELECT 'work_status_ref', work_status_ref, COUNT(*)
FROM events
GROUP BY work_status_ref
UNION ALL
SELECT 'work_category_ref', work_category_ref, COUNT(*)
FROM events
GROUP BY work_category_ref
UNION ALL
SELECT 'road_category', road_category, COUNT(*)
FROM events
GROUP BY road_category
UNION ALL
SELECT 'highway_authority', highway_authority, COUNT(*)
FROM events
GROUP BY highway_authority
UNION ALL
SELECT 'promoter_organisation', promoter_organisation, COUNT(*)
FROM events
GROUP BY promoter_organisation;