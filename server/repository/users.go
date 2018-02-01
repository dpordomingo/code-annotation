package repository

import (
	"database/sql"
	"fmt"

	"github.com/src-d/code-annotation/server/model"
)

// Users repository
type Users struct {
	DB *sql.DB
}

// Create stores a User into the DB, and returns that new User
func (repo *Users) Create(
	login, username, avatarURL string, role model.Role) (*model.User, error) {

	_, err := repo.DB.Exec(
		"INSERT INTO users (login, username, avatar_url, role) VALUES ($1, $2, $3, $4)",
		login, username, avatarURL, role)

	if err != nil {
		return nil, err
	}

	return repo.Get(login)
}

// getWithQuery builds a User from the given sql QueryRow. If the User does not
// exist, it returns nil, nil
func (repo *Users) getWithQuery(queryRow *sql.Row) (*model.User, error) {
	var user model.User

	err := queryRow.Scan(&user.ID, &user.Login, &user.Username, &user.AvatarURL, &user.Role)

	switch {
	case err == sql.ErrNoRows:
		return nil, nil
	case err != nil:
		return nil, fmt.Errorf("Error getting user from the DB: %v", err)
	default:
		return &user, nil
	}
}

// Get returns the User with the given GitHub login name. If the User does not
// exist, it returns nil, nil
func (repo *Users) Get(login string) (*model.User, error) {
	// TODO: escape login string
	return repo.getWithQuery(
		repo.DB.QueryRow("SELECT * FROM users WHERE login=$1", login))
}

// GetByID returns the User with the given ID. If the User does not
// exist, it returns nil, nil
func (repo *Users) GetByID(id int) (*model.User, error) {
	return repo.getWithQuery(
		repo.DB.QueryRow("SELECT * FROM users WHERE id=$1", id))
}
