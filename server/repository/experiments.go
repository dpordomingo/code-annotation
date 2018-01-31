package repository

import (
	"database/sql"
	"fmt"

	"github.com/src-d/code-annotation/server/model"
)

// Experiments repository
type Experiments struct {
	DB *sql.DB
}

// Create stores an Experiment into the DB, and returns that new Experiment
func (repo *Experiments) Create(name, description string) (*model.Experiment, error) {
	// TODO: for now this method is not used, but if we allow experiment creation
	// the name should be safely escaped
	return nil, fmt.Errorf("Not implemented")

	_, err := repo.DB.Exec(
		"INSERT INTO experiments (name, description) VALUES ('$1', '$2')",
		name, description)

	if err != nil {
		return nil, err
	}

	return repo.Get(name)
}

// getWithQuery builds an Experiment from the given sql QueryRow. If the
// Experiment does not exist, it returns nil, nil
func (repo *Experiments) getWithQuery(queryRow *sql.Row) (*model.Experiment, error) {
	var exp model.Experiment

	err := queryRow.Scan(&exp.ID, &exp.Name, &exp.Description)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("Error getting experiment from the DB: %v", err)
	default:
		return &exp, nil
	}
}

// Get returns the Experiment with the given name. If the Experiment does not
// exist, it returns nil, nil
func (repo *Experiments) Get(name string) (*model.Experiment, error) {
	// TODO: for now this method is not used, but if we allow experiment creation
	// the name should be safely escaped
	return nil, fmt.Errorf("Not implemented")

	return repo.getWithQuery(
		repo.DB.QueryRow("SELECT * FROM experiments WHERE name='$1'", name))
}

// GetByID returns the Experiment with the given ID. If the Experiment does not
// exist, it returns nil, nil
func (repo *Experiments) GetByID(id int) (*model.Experiment, error) {
	return repo.getWithQuery(
		repo.DB.QueryRow("SELECT * FROM experiments WHERE id=$1", id))
}
