package internal

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"strings"

	"github.com/cockroachdb/errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rm-hull/street-manager-relay/models"
	"github.com/tavsec/gin-healthcheck/checks"
)

//go:embed sql/create_db.sql
var createSQL string

//go:embed sql/search.sql
var searchSQL string

//go:embed sql/ref_data.sql
var refDataSQL string

type DbRepository struct {
	db          *sql.DB
	searchStmt  *sql.Stmt
	refDataStmt *sql.Stmt
}

type Batch struct {
	tx   *sql.Tx
	stmt *sql.Stmt
}

func NewDbRepository(dbPath string) (*DbRepository, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	if err = db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	err = create(db)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create database")
	}

	searchStmt, err := db.Prepare(searchSQL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare search SQL")
	}

	refDataStmt, err := db.Prepare(refDataSQL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare ref-data SQL")
	}

	log.Printf("Database initialized successfully: %s", dbPath)
	return &DbRepository{
		db:          db,
		searchStmt:  searchStmt,
		refDataStmt: refDataStmt,
	}, nil
}

func create(db *sql.DB) error {
	exists, err := tablesExists(db, "events")
	if err != nil {
		return errors.Wrap(err, "error checking if table exists")
	}
	if exists {
		return nil
	}
	_, err = db.Exec(createSQL)
	return err
}

func tablesExists(db *sql.DB, table string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)"
	err := db.QueryRow(query, table).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func (repo *DbRepository) RefData() (*models.RefData, error) {
	refData := make(models.RefData)

	rows, err := repo.refDataStmt.Query()
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute refData query")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var facet string
	var value sql.NullString
	var count int

	for rows.Next() {
		if err := rows.Scan(&facet, &value, &count); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		if _, ok := refData[facet]; !ok {
			refData[facet] = make(map[string]int)
		}
		v := ""
		if value.Valid {
			v = value.String
		}
		refData[facet][v] = count
	}
	return &refData, nil
}

func (repo *DbRepository) Search(bbox *models.BBox, facets *models.Facets, temporalFilters *models.TemporalFilters) ([]*models.Event, error) {
	if bbox == nil {
		return nil, errors.New("bounding box is required")
	}

	params := facetsToParams(bbox, facets, temporalFilters)
	rows, err := repo.searchStmt.Query(params...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute search query")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	events := make([]*models.Event, 0, 50)
	for rows.Next() {
		var event models.Event
		if err := rows.Scan(
			// Identifiers
			&event.ID,
			&event.EventType,
			&event.ObjectReference,
			&event.ActivityReferenceNumber,
			&event.WorkReferenceNumber,
			&event.Section58ReferenceNumber,
			&event.PermitReferenceNumber,

			// Core location and authority info
			&event.USRN,
			&event.StreetName,
			&event.AreaName,
			&event.Town,
			&event.HighwayAuthority,
			&event.HighwayAuthoritySWACode,
			&event.PromoterSWACode,
			&event.PromoterOrganisation,

			// Coordinates & descriptions
			&event.ActivityCoordinates,
			&event.ActivityLocationType,
			&event.ActivityLocationDescription,
			&event.WorksLocationCoordinates,
			&event.WorksLocationType,
			&event.Section58Coordinates,
			&event.Section58LocationType,

			// Categories & types
			&event.WorkCategory,
			&event.WorkCategoryRef,
			&event.WorkStatus,
			&event.WorkStatusRef,
			&event.TrafficManagementType,
			&event.TrafficManagementTypeRef,
			&event.CurrentTrafficManagementType,
			&event.CurrentTrafficManagementTypeRef,
			&event.RoadCategory,
			&event.ActivityType,
			&event.ActivityTypeDetails,
			&event.Section58Status,
			&event.Section58Duration,
			&event.Section58Extent,

			// Dates/times
			&event.ProposedStartDate,
			&event.ProposedEndDate,
			&event.ProposedStartTime,
			&event.ProposedEndTime,
			&event.ActualStartDateTime,
			&event.ActualEndDateTime,
			&event.StartDate,
			&event.StartTime,
			&event.EndDate,
			&event.EndTime,
			&event.CurrentTrafficManagementUpdateDate,

			// Flags
			&event.IsTTRORequired,
			&event.IsCovid19Response,
			&event.IsTrafficSensitive,
			&event.IsDeemed,
			&event.CollaborativeWorking,
			&event.Cancelled,
			&event.TrafficManagementRequired,

			// Misc
			&event.PermitConditions,
			&event.PermitStatus,
			&event.CollaborationType,
			&event.CollaborationTypeRef,
			&event.CloseFootway,
			&event.CloseFootwayRef,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}
		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating over rows")
	}

	return events, nil
}

func facetsToParams(bbox *models.BBox, facets *models.Facets, temporalFilters *models.TemporalFilters) []any {
	params := []any{
		// Correct parameter order: maxX, minX, maxY, minY (which seems counterintuitive)
		bbox.MaxX, bbox.MinX, bbox.MaxY, bbox.MinY,
		fmt.Sprintf("+%d days", temporalFilters.MaxDaysAhead),
		fmt.Sprintf("-%d days", temporalFilters.MaxDaysBehind),
	}

	// Add string facet parameters (each facet needs 3 parameters for the OR condition)
	facetGetters := []func(f *models.Facets) []string{
		func(f *models.Facets) []string { return f.PermitStatus },
		func(f *models.Facets) []string { return f.TrafficManagementTypeRef },
		func(f *models.Facets) []string { return f.WorkStatusRef },
		func(f *models.Facets) []string { return f.WorkCategoryRef },
		func(f *models.Facets) []string { return f.RoadCategory },
		func(f *models.Facets) []string { return f.HighwayAuthority },
		func(f *models.Facets) []string { return f.PromoterOrganisation },
	}

	for _, getter := range facetGetters {
		jsonVal := toJSONOrNil(getFacetSlice(facets, getter))
		// Needs to be copied 3x times for the placeholders
		params = append(params, jsonVal, jsonVal, jsonVal)
	}

	return params
}

// Helper function to convert string slice to JSON or nil
func toJSONOrNil(slice []string) any {
	if len(slice) == 0 {
		return nil
	}
	jsonBytes, _ := json.Marshal(slice)
	return string(jsonBytes)
}

// Helper functions to safely extract facet values
func getFacetSlice(facets *models.Facets, getter func(*models.Facets) []string) []string {
	if facets == nil {
		return nil
	}
	return getter(facets)
}

func (repo *DbRepository) Close() error {
	if repo.searchStmt != nil {
		if err := repo.searchStmt.Close(); err != nil {
			return errors.Wrap(err, "failed to close search db statement")
		}
	}

	if repo.refDataStmt != nil {
		if err := repo.refDataStmt.Close(); err != nil {
			return errors.Wrap(err, "failed to close ref-data db statement")
		}
	}

	if repo.db != nil {
		return repo.db.Close()
	}
	return nil
}

func (repo *DbRepository) HealthCheck() checks.Check {
	return checks.SqlCheck{
		Sql: repo.db,
	}
}

func (repo *DbRepository) BatchUpsert() (*Batch, error) {
	cols := []string{
		// Identifiers
		"event_type",
		"object_reference",
		"activity_reference_number",
		"work_reference_number",
		"section_58_reference_number",
		"permit_reference_number",

		// Core location and authority info
		"usrn",
		"street_name",
		"area_name",
		"town",
		"highway_authority",
		"highway_authority_swa_code",
		"promoter_swa_code",
		"promoter_organisation",

		// Coordinates & descriptions
		"activity_coordinates",
		"activity_location_type",
		"activity_location_description",
		"works_location_coordinates",
		"works_location_type",
		"section_58_coordinates",
		"section_58_location_type",

		// Categories & types
		"work_category",
		"work_category_ref",
		"work_status",
		"work_status_ref",
		"traffic_management_type",
		"traffic_management_type_ref",
		"current_traffic_management_type",
		"current_traffic_management_type_ref",
		"road_category",
		"activity_type",
		"activity_type_details",
		"section_58_status",
		"section_58_duration",
		"section_58_extent",

		// Dates/times
		"proposed_start_date",
		"proposed_end_date",
		"proposed_start_time",
		"proposed_end_time",
		"actual_start_date_time",
		"actual_end_date_time",
		"start_date",
		"start_time",
		"end_date",
		"end_time",
		"current_traffic_management_update_date",

		// Flags
		"is_ttro_required",
		"is_covid_19_response",
		"is_traffic_sensitive",
		"is_deemed",
		"collaborative_working",
		"cancelled",
		"traffic_management_required",

		// Misc
		"permit_conditions",
		"permit_status",
		"collaboration_type",
		"collaboration_type_ref",
		"close_footway",
		"close_footway_ref",
	}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = "?"
	}

	// Build the ON CONFLICT update set clauses
	updateSet := make([]string, len(cols)-1) // exclude the unique key itself
	for i, col := range cols[1:] {
		updateSet[i] = fmt.Sprintf("%s=excluded.%s", col, col)
	}

	query := fmt.Sprintf(`
		INSERT INTO events (%s)
		VALUES (%s)
		ON CONFLICT(object_reference) DO UPDATE SET
		%s
		RETURNING id;
	`, strings.Join(cols, ", "), strings.Join(placeholders, ", "), strings.Join(updateSet, ", "))

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "failed to begin transaction")
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare statement")
	}

	return &Batch{
		tx:   tx,
		stmt: stmt,
	}, nil
}

// inserts or updates an event based on object reference.
func (batch *Batch) Upsert(event *models.Event) (int64, error) {
	bbox, err := event.BoundingBox()
	if err != nil {
		return 0, errors.Wrap(err, "failed to calculate bounding box")
	}

	// Extract values from struct
	values := []any{
		// Identifiers
		event.EventType,
		event.ObjectReference,
		event.ActivityReferenceNumber,
		event.WorkReferenceNumber,
		event.Section58ReferenceNumber,
		event.PermitReferenceNumber,

		// Core location and authority info
		event.USRN,
		event.StreetName,
		event.AreaName,
		event.Town,
		event.HighwayAuthority,
		event.HighwayAuthoritySWACode,
		event.PromoterSWACode,
		event.PromoterOrganisation,

		// Coordinates & descriptions
		event.ActivityCoordinates,
		event.ActivityLocationType,
		event.ActivityLocationDescription,
		event.WorksLocationCoordinates,
		event.WorksLocationType,
		event.Section58Coordinates,
		event.Section58LocationType,

		// Categories & types
		event.WorkCategory,
		event.WorkCategoryRef,
		event.WorkStatus,
		event.WorkStatusRef,
		event.TrafficManagementType,
		event.TrafficManagementTypeRef,
		event.CurrentTrafficManagementType,
		event.CurrentTrafficManagementTypeRef,
		event.RoadCategory,
		event.ActivityType,
		event.ActivityTypeDetails,
		event.Section58Status,
		event.Section58Duration,
		event.Section58Extent,

		// Dates/times
		event.ProposedStartDate,
		event.ProposedEndDate,
		event.ProposedStartTime,
		event.ProposedEndTime,
		event.ActualStartDateTime,
		event.ActualEndDateTime,
		event.StartDate,
		event.StartTime,
		event.EndDate,
		event.EndTime,
		event.CurrentTrafficManagementUpdateDate,

		// Flags
		event.IsTTRORequired,
		event.IsCovid19Response,
		event.IsTrafficSensitive,
		event.IsDeemed,
		event.CollaborativeWorking,
		event.Cancelled,
		event.TrafficManagementRequired,

		// Misc
		event.PermitConditions,
		event.PermitStatus,
		event.CollaborationType,
		event.CollaborationTypeRef,
		event.CloseFootway,
		event.CloseFootwayRef,
	}

	var id int64
	err = batch.stmt.QueryRow(values...).Scan(&id)
	if err != nil {
		return 0, errors.Wrap(err, "failed to execute upsert query")
	}

	err = batch.upsertRTree(id, *bbox)
	if err != nil {
		return 0, errors.Wrap(err, "failed to insert into R-tree")
	}

	return id, nil
}

func (batch *Batch) Done() error {
	if commitErr := batch.tx.Commit(); commitErr != nil {
		return errors.Wrap(commitErr, "failed to commit transaction")
	}

	return nil
}

func (batch *Batch) Abort(err error) error {
	if rbErr := batch.tx.Rollback(); rbErr != nil {
		return errors.Wrapf(rbErr, "failed to rollback transaction: original error: %w", err)
	}
	return err
}

func (batch *Batch) upsertRTree(id int64, bbox models.BBox) error {
	_, err := batch.tx.Exec(`DELETE FROM events_rtree WHERE id = ?`, id)
	if err != nil {
		return err
	}

	_, err = batch.tx.Exec(
		`INSERT INTO events_rtree (id, minx, maxx, miny, maxy) VALUES (?, ?, ?, ?, ?)`,
		id, bbox.MinX, bbox.MaxX, bbox.MinY, bbox.MaxY,
	)
	return err
}

func (repo *DbRepository) RegenerateIndex() (int, int, error) {

	tx, err := repo.db.Begin()
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to begin transaction")
	}

	defer func() {
		// Rollback will have no effect if commit was already called
		if err := tx.Rollback(); err != nil {
			log.Printf("Rollback: %v", err)
		}
	}()

	updateStmt, err := tx.Prepare(`
		UPDATE events_rtree
		SET minx=?, maxx=?, miny=?, maxy=?
		WHERE ID=?
	`)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to prepare update statement")
	}

	rows, err := repo.db.Query(`
		SELECT
			e.id,
			COALESCE(e.works_location_coordinates, e.activity_coordinates, e.section_58_coordinates) as coords,
			r.minx,
			r.maxx,
			r.miny,
			r.maxy
		FROM events e
		INNER JOIN events_rtree r ON e.id = r.id
	`)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to execute query")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	affected := 0
	total := 0

	var id int64
	var coords string
	var bbox models.BBox

	for rows.Next() {
		if err := rows.Scan(&id, &coords, &bbox.MinX, &bbox.MaxX, &bbox.MinY, &bbox.MaxY); err != nil {
			return 0, 0, errors.Wrap(err, "failed to scan row")
		}

		regen, err := models.BoundingBoxFromWKT(coords)
		if err != nil {
			return 0, 0, errors.Wrap(err, "failed to create bounding box")
		}

		if !regen.Equals(bbox, 1) {
			log.Printf("Record %d needs bbox regen: %v doesnt match stored: %v", id, regen, bbox)
			affected++

			res, err := updateStmt.Exec(regen.MinX, regen.MaxX, regen.MinY, regen.MaxY, id)
			if err != nil {
				return 0, 0, errors.Wrap(err, "failed to update rtree")
			}

			updated, err := res.RowsAffected()
			if err != nil {
				return 0, 0, errors.Wrap(err, "failed to get rows affected")
			}
			if updated != 1 {
				return 0, 0, errors.Newf("unexpected rows affected %d for id=%d", updated, id)
			}
		}

		total++
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, errors.Wrap(err, "failed to commit transaction")
	}

	return affected, total, nil

}
