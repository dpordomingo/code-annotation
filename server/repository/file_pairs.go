package repository

import (
	"database/sql"
	"fmt"

	"github.com/src-d/code-annotation/server/model"
)

// FilePairs repository
type FilePairs struct {
	DB *sql.DB
}

// getWithQuery builds a FilePair from the given sql QueryRow. If the FilePair
// does not exist, it returns nil, nil
func (repo *FilePairs) getWithQuery(queryRow *sql.Row) (*model.FilePair, error) {
	var pair model.FilePair

	var a, b model.File

	err := queryRow.Scan(&pair.ID,
		&a.BlobID, &a.RepositoryID, &a.CommitHash, &a.Path, &a.Content, &a.Hash,
		&b.BlobID, &b.RepositoryID, &b.CommitHash, &b.Path, &b.Content, &b.Hash,
		&pair.Score, &pair.Diff, &pair.ExperimentID)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("Error getting file pair from the DB: %v", err)
	default:
		return &pair, nil
	}
}

// GetByID returns the FilePair with the given ID. If the FilePair does not
// exist, it returns nil, nil
func (repo *FilePairs) GetByID(id int) (*model.FilePair, error) {
	return repo.getWithQuery(
		repo.DB.QueryRow("SELECT * FROM file_pairs WHERE id=$1", id))
}
