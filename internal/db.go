package internal

import (
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rm-hull/street-manager-relay/models"
)

//go:embed create_db.sql
var createSQL string

type DbRepository struct {
	db *sql.DB
}

func NewDbRepository(dbPath string) (*DbRepository, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	repo := &DbRepository{db: db}
	err = repo.create()
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	log.Printf("Database initialized successfully: %s", dbPath)
	return repo, nil
}

func (repo *DbRepository) create() error {
	exists, err := repo.tablesExists("activities")
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
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='%s')", table)
	err := repo.db.QueryRow(query).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

func (repo *DbRepository) Close() error {
	if repo.db != nil {
		return repo.db.Close()
	}
	return nil
}

// inserts or updates an Activity based on activity_reference_number.
func (repo *DbRepository) Upsert(activity *models.Activity) (int64, error) {
	if activity.ActivityReferenceNumber == nil || *activity.ActivityReferenceNumber == "" {
		return 0, fmt.Errorf("activity_reference_number is required")
	}

	bbox, err := activity.BoundingBox()
	if err != nil {
		return 0, fmt.Errorf("failed to calculate bounding box: %w", err)
	}

	cols := []string{
		"activity_reference_number",
		"usrn",
		"street_name",
		"area_name",
		"town",
		"highway_authority",
		"highway_authority_swa_code",
		"activity_coordinates",
		"activity_location_type",
		"activity_location_description",
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
		"is_ttro_required",
		"is_covid_19_response",
		"is_traffic_sensitive",
		"is_deemed",
		"collaborative_working",
		"cancelled",
		"traffic_management_required",
		"works_location_coordinates",
		"works_location_type",
		"permit_conditions",
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
		INSERT INTO activities (%s)
		VALUES (%s)
		ON CONFLICT(activity_reference_number) DO UPDATE SET
		%s
	`, strings.Join(cols, ", "), strings.Join(placeholders, ", "), strings.Join(updateSet, ", "))

	// Extract values from struct
	values := []any{
		activity.ActivityReferenceNumber,
		activity.USRN,
		activity.StreetName,
		activity.AreaName,
		activity.Town,
		activity.HighwayAuthority,
		activity.HighwayAuthoritySwaCode,
		activity.ActivityCoordinates,
		activity.ActivityLocationType,
		activity.ActivityLocationDescription,
		activity.WorkCategory,
		activity.WorkCategoryRef,
		activity.WorkStatus,
		activity.WorkStatusRef,
		activity.TrafficManagementType,
		activity.TrafficManagementTypeRef,
		activity.CurrentTrafficManagementType,
		activity.CurrentTrafficManagementTypeRef,
		activity.RoadCategory,
		activity.ActivityType,
		activity.ActivityTypeDetails,
		activity.ProposedStartDate,
		activity.ProposedEndDate,
		activity.ProposedStartTime,
		activity.ProposedEndTime,
		activity.ActualStartDateTime,
		activity.ActualEndDateTime,
		activity.StartDate,
		activity.StartTime,
		activity.EndDate,
		activity.EndTime,
		activity.CurrentTrafficManagementUpdateDate,
		activity.IsTtroRequired,
		activity.IsCovid19Response,
		activity.IsTrafficSensitive,
		activity.IsDeemed,
		activity.CollaborativeWorking,
		activity.Cancelled,
		activity.TrafficManagementRequired,
		activity.WorksLocationCoordinates,
		activity.WorksLocationType,
		activity.PermitConditions,
		activity.CollaborationType,
		activity.CollaborationTypeRef,
		activity.CloseFootway,
		activity.CloseFootwayRef,
	}

	tx, err := repo.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}

	res, err := tx.Exec(query, values...)
	if err != nil {
		return 0, tryRollback(fmt.Errorf("failed to execute upsert query: %w", err), tx)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, tryRollback(fmt.Errorf("failed to get last insert id: %w", err), tx)
	}

	err = upsertRTree(tx, id, *bbox)
	if err != nil {
		return 0, tryRollback(fmt.Errorf("failed to insert into R-tree: %w", err), tx)
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func tryRollback(err error, tx *sql.Tx) error {
	if rbErr := tx.Rollback(); rbErr != nil {
		return fmt.Errorf("failed to rollback transaction: %v; original error: %w", rbErr, err)
	}
	return err
}

func upsertRTree(tx *sql.Tx, id int64, bbox models.BBox) error {
	_, err := tx.Exec(`DELETE FROM activities_rtree WHERE id = ?`, id)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO activities_rtree (id, minx, maxx, miny, maxy) VALUES (?, ?, ?, ?, ?)`,
		id, bbox.MinX, bbox.MaxX, bbox.MinY, bbox.MaxY,
	)
	return err
}
