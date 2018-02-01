package repository

import (
	"database/sql"
	"fmt"

	"github.com/src-d/code-annotation/server/model"
)

// Assignments repository
type Assignments struct {
	DB *sql.DB
}

// ErrNoAssignmentsInitialized is the error returned when the Assignments of a
// User are requested for a given Experiment, but they have not been yet created
var ErrNoAssignmentsInitialized = fmt.Errorf("No assignments initialized")

// Initialize builds the assignments for the given user and experiment IDs
func (repo *Assignments) Initialize(userID int, experimentID int) ([]*model.Assignment, error) {
	tx, err := repo.DB.Begin()
	if err != nil {
		return nil, err
	}

	insert, err := tx.Prepare("INSERT INTO assignments VALUES ($1, $2, $3, $4, $5)")
	if err != nil {
		return nil, fmt.Errorf("DB error: %v", err)
	}

	rows, err := repo.DB.Query(
		"SELECT id FROM file_pairs WHERE experiment_id=$1", experimentID)
	if err != nil {
		return nil, fmt.Errorf("Error getting file_pairs from the DB: %v", err)
	}

	answer := ""
	duration := 0

	for rows.Next() {
		var pairID int
		rows.Scan(&pairID)

		_, err := insert.Exec(userID, pairID, experimentID, answer, duration)
		if err != nil {
			return nil, fmt.Errorf("DB error: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("DB error: %v", err)
	}

	return repo.GetAll(userID, experimentID)
}

// Create stores an Assignment into the DB, and returns that new Assignment
func (repo *Assignments) Create(as *model.Assignment) error {

	_, err := repo.DB.Exec(
		`INSERT INTO assignments (user_id, pair_id, experiment_id, answer, duration)
		VALUES ($1, $2, $3, $4, $5)`,
		as.UserID, as.PairID, as.ExperimentID, as.Answer, as.Duration)

	if err != nil {
		return err
	}

	as, err = repo.Get(as.UserID, as.PairID)
	return err
}

// getWithQuery builds a Assignment from the given sql QueryRow. If the
// Assignment does not exist, it returns nil, nil
func (repo *Assignments) getWithQuery(queryRow *sql.Row) (*model.Assignment, error) {
	var as model.Assignment

	err := queryRow.Scan(
		&as.UserID, &as.PairID, &as.ExperimentID, &as.Answer, &as.Duration)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("Error getting assignment from the DB: %v", err)
	default:
		return &as, nil
	}
}

// Get returns the Assignment for the given user and pair IDs. If the Assignment
// does not exist, it returns nil, nil
func (repo *Assignments) Get(userID, pairID int) (*model.Assignment, error) {
	return repo.getWithQuery(repo.DB.QueryRow(
		"SELECT * FROM assignments WHERE user_id=$1 AND pair_id=$2",
		userID, pairID))
}

// GetAll returns all the Assignments for the given user and experiment IDs.
// Returns an ErrNoAssignmentsInitialized if they do not exist yet
func (repo *Assignments) GetAll(userID, experimentID int) ([]*model.Assignment, error) {
	rows, err := repo.DB.Query(
		"SELECT * FROM assignments WHERE user_id=$1 AND experiment_id=$2",
		userID, experimentID)
	if err != nil {
		return nil, fmt.Errorf("Error getting assignments from the DB: %v", err)
	}

	results := make([]*model.Assignment, 0)

	for rows.Next() {
		var as model.Assignment
		rows.Scan(
			&as.UserID, &as.PairID, &as.ExperimentID, &as.Answer, &as.Duration)

		results = append(results, &as)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DB error: %v", err)
	}

	if len(results) == 0 {
		return nil, ErrNoAssignmentsInitialized
	}

	return results, nil
}

// Update updates the Assignment identified by the given user and pair IDs,
// with the given answer and duration
func (repo *Assignments) Update(userID int, pairID int, answer string, duration int) error {
	if _, ok := model.Answers[answer]; !ok {
		return fmt.Errorf("Wrong answer provided: '%s'", answer)
	}

	cmd := fmt.Sprintf(
		"UPDATE assignments SET answer='%v', duration=%v WHERE user_id=%v AND pair_id=%v",
		answer, duration, userID, pairID)

	_, err := repo.DB.Exec(cmd)

	return err
}
