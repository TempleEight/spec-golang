package dao

import (
	"database/sql"
	"fmt"

	"github.com/TempleEight/spec-golang/user/util"
	// pq acts as the driver for SQL requests
	_ "github.com/lib/pq"
)

// Datastore provides the interface adopted by the DAO, allowing for mocking
type Datastore interface {
	CreateUser(input CreateUserInput) (*User, error)
	ReadUser(input ReadUserInput) (*User, error)
	UpdateUser(input UpdateUserInput) (*User, error)
	DeleteUser(input DeleteUserInput) error
}

// DAO encapsulates access to the database
type DAO struct {
	DB *sql.DB
}

// User encapsulates the object stored in the datastore
type User struct {
	ID     int64
	AuthID int64
	Name   string
}

// CreateUserInput encapsulates the information required to create a single user
type CreateUserInput struct {
	AuthID int64
	Name   string
}

// ReadUserInput encapsulates the information required to read a single user
type ReadUserInput struct {
	ID int64
}

// UpdateUserInput encapsulates the information required to update a single user
type UpdateUserInput struct {
	ID   int64
	Name string
}

// DeleteUserInput encapsulates the information required to delete a single user
type DeleteUserInput struct {
	ID int64
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
func (dao *DAO) CreateUser(input CreateUserInput) (*User, error) {
	row := executeQueryWithRowResponse(dao.DB, "INSERT INTO User_Temple (auth_id, name) VALUES ($1, $2) RETURNING *", input.AuthID, input.Name)

	var user User
	err := row.Scan(&user.ID, &user.AuthID, &user.Name)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ReadUser returns the information about a user stored for a given ID
func (dao *DAO) ReadUser(input ReadUserInput) (*User, error) {
	row := executeQueryWithRowResponse(dao.DB, "SELECT * FROM User_Temple WHERE id = $1", input.ID)

	var user User
	err := row.Scan(&user.ID, &user.Name)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrUserNotFound(input.ID)
		default:
			return nil, err
		}
	}

	return &user, nil
}

// UpdateUser updates a user in the database, returning an error if it fails
func (dao *DAO) UpdateUser(input UpdateUserInput) (*User, error) {
	row := executeQueryWithRowResponse(dao.DB, "UPDATE User_Temple set Name = $1 WHERE Id = $2 RETURNING *", input.Name, input.ID)

	var user User
	err := row.Scan(&user.ID, &user.AuthID, &user.Name)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrUserNotFound(input.ID)
		default:
			return nil, err
		}
	}

	return &user, nil
}

// DeleteUser deletes a user in the database, returning an error if it fails or the user doesn't exist
func (dao *DAO) DeleteUser(input DeleteUserInput) error {
	rowsAffected, err := executeQuery(dao.DB, "DELETE FROM User_Temple WHERE Id = $1", input.ID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrUserNotFound(input.ID)
	}

	return nil
}
