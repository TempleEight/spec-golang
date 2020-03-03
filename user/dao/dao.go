package dao

import (
	"database/sql"
	"fmt"

	"github.com/TempleEight/spec-golang/user/util"
	// pq acts as the driver for SQL requests
	_ "github.com/lib/pq"
)

type Datastore interface {
	CreateUser(request UserCreateRequest) (*UserCreateResponse, error)
	ReadUser(userID int64) (*UserReadResponse, error)
	UpdateUser(userID int64, request UserUpdateRequest) (*UserUpdateResponse, error)
	DeleteUser(userID int64) error
}

// DAO encapsulates access to the database
type DAO struct {
	DB *sql.DB
}

// UserCreateRequest contains the information required to create a new user
type UserCreateRequest struct {
	Name string `valid:"type(string),required,stringlength(2|255)"`
}

// UserUpdateRequest contains all the information about an existing user
type UserUpdateRequest struct {
	Name string `valid:"type(string),required,stringlength(2|255)"`
}

// UserCreateResponse contains the information about the newly created user
type UserCreateResponse struct {
	ID   int
	Name string
}

// UserReadResponse returns all the information stored about a user
type UserReadResponse struct {
	ID   int
	Name string
}

// UserUpdateResponse contains the information about the newly updated user
type UserUpdateResponse struct {
	ID   int
	Name string
}

// Executes the query, returning the row
func executeQueryWithRowResponse(db *sql.DB, query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...)
}

// Executes a query, returning the number of rows affected
func executeQuery(db *sql.DB, query string, args ...interface{}) (int64, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Init opens the database connection
func Init(config *util.Config) (*DAO, error) {
	connStr := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=%s", config.User, config.DBName, config.Host, config.SSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &DAO{db}, nil
}

// CreateUser creates a new user in the database, returning the newly created user
func (dao *DAO) CreateUser(request UserCreateRequest) (*UserCreateResponse, error) {
	row := executeQueryWithRowResponse(dao.DB, "INSERT INTO User_Temple (name) VALUES ($1) RETURNING *", request.Name)

	var resp UserCreateResponse
	err := row.Scan(&resp.ID, &resp.Name)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ReadUser returns the information about a user stored for a given ID
func (dao *DAO) ReadUser(userID int64) (*UserReadResponse, error) {
	row := executeQueryWithRowResponse(dao.DB, "SELECT * FROM User_Temple WHERE id = $1", userID)

	var resp UserReadResponse
	err := row.Scan(&resp.ID, &resp.Name)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrUserNotFound(userID)
		default:
			return nil, err
		}
	}

	return &resp, nil
}

// UpdateUser updates a user in the database, returning an error if it fails
func (dao *DAO) UpdateUser(userID int64, request UserUpdateRequest) (*UserUpdateResponse, error) {
	row := executeQueryWithRowResponse(dao.DB, "UPDATE User_Temple set Name = $1 WHERE Id = $2 RETURNING *", request.Name, userID)

	var resp UserUpdateResponse
	err := row.Scan(&resp.ID, &resp.Name)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrUserNotFound(userID)
		default:
			return nil, err
		}
	}

	return &resp, nil
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
