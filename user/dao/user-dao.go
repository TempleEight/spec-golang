package dao

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/TempleEight/spec-golang/user/utils"
	// pq acts as the driver for SQL requests
	_ "github.com/lib/pq"
)

// DAO encapsulates access to the database
type DAO struct {
	DB *sql.DB
}

// UserGetResponse returns all the information stored about a user
type UserGetResponse struct {
	ID   int
	Name string
}

// UserCreateRequest contains the information required to create a new user
type UserCreateRequest struct {
	Name string `valid:"type(string),required,stringlength(2|255)"`
}

// UserCreateResponse contains the information stored about the newly created user
type UserCreateResponse struct {
	ID   int
	Name string
}

// UserUpdateRequest contains all the information about an existing user
type UserUpdateRequest struct {
	Name string `valid:"type(string),required,stringlength(2|255)"`
}

// Executes the query, returning the row
func executeQueryWithRowResponse(db *sql.DB, query string, args ...interface{}) (*sql.Row, error) {
	return db.QueryRow(query, args...), nil
}

// Executes a query, returning the number of rows affected
func executeQuery(db *sql.DB, query string, args ...interface{}) (int64, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Initialise opens the database connection
func (dao *DAO) Initialise(config *utils.Config) error {
	connStr := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=%s", config.User, config.DBName, config.Host, config.SSLMode)
	var err error
	dao.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	return nil
}

// GetUser returns the information about a user stored for a given ID
func (dao *DAO) GetUser(id int64) (*UserGetResponse, error) {
	row, err := executeQueryWithRowResponse(dao.DB, "SELECT * FROM User_Temple WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	var user UserGetResponse
	err = row.Scan(&user.ID, &user.Name)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrUserNotFound(id)
		default:
			return nil, err
		}
	}

	return &user, nil
}

// CreateUser creates a new user in the database, returning the newly created user
func (dao *DAO) CreateUser(request UserCreateRequest) (*UserCreateResponse, error) {
	row, err := executeQueryWithRowResponse(dao.DB, "INSERT INTO User_Temple (name) VALUES ($1) RETURNING *", request.Name)
	if err != nil {
		return nil, err
	}

	var user UserCreateResponse
	err = row.Scan(&user.ID, &user.Name)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("Could not create user")
		default:
			return nil, err
		}
	}

	return &user, nil
}

// UpdateUser updates a user in the database, returning an error if it fails
func (dao *DAO) UpdateUser(userID int64, request UserUpdateRequest) error {
	query := "UPDATE User_Temple set Name = $1 WHERE Id = $2"
	rowsAffected, err := executeQuery(dao.DB, query, request.Name, userID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrUserNotFound(userID)
	}

	return nil
}

// DeleteUser deletes a user in the database, returning an error if it fails or the user doesn't exist
func (dao *DAO) DeleteUser(userID int64) error {
	rowsAffected, err := executeQuery(dao.DB, "DELETE FROM User_Temple WHERE Id = $1", userID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrUserNotFound(userID)
	}

	return nil
}
