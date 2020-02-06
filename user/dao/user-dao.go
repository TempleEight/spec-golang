package dao

import (
	"database/sql"
	"fmt"

	// pq acts as the driver for SQL requests
	_ "github.com/lib/pq"
)

// UserGetResponse returns all the information stored about a user
type UserGetResponse struct {
	ID   int
	Name string
}

// GetUser returns the information about a user stored for a given ID
func GetUser(id int64) (*UserGetResponse, error) {
	// Host matches the container name in docker-compose.yml
	// https://docs.docker.com/compose/networking/
	connStr := "user=postgres dbname=postgres host=user-db sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var user UserGetResponse
	err = db.QueryRow("SELECT * FROM Users WHERE id = $1", id).Scan(&user.ID, &user.Name)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, fmt.Errorf("User for ID %d does not exist", id)
		default:
			return nil, err
		}
	}
	return &user, nil
}
