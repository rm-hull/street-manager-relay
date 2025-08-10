package internal

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rm-hull/street-manager-relay/models"
	"github.com/tavsec/gin-healthcheck/checks"
)

//go:embed sql/create_db.sql
var createSQL string

//go:embed sql/search.sql
var searchSQL string

type DbRepository struct {
	db         *sql.DB
	searchStmt *sql.Stmt
}

type Batch struct {
	tx   *sql.Tx
	stmt *sql.Stmt
}

func NewDbRepository(dbPath string) (*DbRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	searchStmt, err := db.Prepare(searchSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare search SQL: %w", err)
	}

	repo := &DbRepository{db: db, searchStmt: searchStmt}
	err = repo.create()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("Database initialized successfully: %s", dbPath)
	return repo, nil
}

func (repo *DbRepository) create() error {
	exists, err := repo.tablesExists("events")
	if err != nil {
		return fmt.Errorf("error checking if table exists: %w", err)
	}
	if exists {
		return nil
	}
	_, err = repo.db.Exec(createSQL)
	return err
}

func (repo *DbRepository) tablesExists(table string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name=?)"
	err := repo.db.QueryRow(query, table).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func (repo *DbRepository) Search(bbox *models.BBox) ([]models.Event, error) {
	if bbox == nil {
		return nil, fmt.Errorf("bounding box is required")
	}

	// Correct parameter order: maxX, minX, maxY, minY (which seems counterintuitive)
	rows, err := repo.searchStmt.Query(bbox.MaxX, bbox.MinX, bbox.MaxY, bbox.MinY)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var events []models.Event
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
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return events, nil
}

func (repo *DbRepository) Close() error {
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
	`, strings.Join(cols, ", "), strings.Join(placeholders, ", "), strings.Join(updateSet, ", "))

	tx, err := repo.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
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
		return 0, fmt.Errorf("failed to calculate bounding box: %w", err)
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

	res, err := batch.stmt.Exec(values...)
	if err != nil {
		return 0, fmt.Errorf("failed to execute upsert query: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	err = batch.upsertRTree(id, *bbox)
	if err != nil {
		return 0, fmt.Errorf("failed to insert into R-tree: %w", err)
	}

	return id, nil
}

func (batch *Batch) Done() error {
	if commitErr := batch.tx.Commit(); commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}

	return nil
}

func (batch *Batch) Abort(err error) error {
	if rbErr := batch.tx.Rollback(); rbErr != nil {
		return fmt.Errorf("failed to rollback transaction: %v; original error: %w", rbErr, err)
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
