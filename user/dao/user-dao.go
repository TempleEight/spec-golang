package dao

import (
	"database/sql"
	"errors"

	// pq acts as the driver for SQL requests
	_ "github.com/lib/pq"
)

// TODO: This should come from a configuration file
// Host matches the container name in docker-compose.yml
// https://docs.docker.com/compose/networking/
const connStr = "user=postgres dbname=postgres host=user-db sslmode=disable"

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

func executeQueryWithRowResponse(query string, args ...interface{}) (*sql.Row, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return db.QueryRow(query, args...), nil
}

// Executes a query, returning the number of rows affected
func executeQuery(query string, args ...interface{}) (int64, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// GetUser returns the information about a user stored for a given ID
func GetUser(id int64) (*UserGetResponse, error) {
	row, err := executeQueryWithRowResponse("SELECT * FROM Users WHERE id = $1", id)
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
func CreateUser(request UserCreateRequest) (*UserCreateResponse, error) {
	row, err := executeQueryWithRowResponse("INSERT INTO Users (name) VALUES ($1) RETURNING *", request.Name)
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
func UpdateUser(userID int64, request UserUpdateRequest) error {
	query := "UPDATE Users set Name = $1 WHERE Id = $2"
	rowsAffected, err := executeQuery(query, request.Name, userID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrUserNotFound(userID)
	}

	return nil
}

// DeleteUser deletes a user in the database, returning an error if it fails or the user doesn't exist
func DeleteUser(userID int64) error {
	rowsAffected, err := executeQuery("DELETE FROM Users WHERE Id = $1", userID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrUserNotFound(userID)
	}

	return nil
}
