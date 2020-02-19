package dao

import (
	"database/sql"
	"fmt"

	"github.com/TempleEight/spec-golang/match/utils"
	// pq acts as the driver for SQL requests
	_ "github.com/lib/pq"
)

// DAO encapsulates access to the database
type DAO struct {
	DB *sql.DB
}

// MatchCreateRequest contains the information required to create a new match
type MatchCreateRequest struct {
	UserOne *int `valid:"-"`
	UserTwo *int `valid:"-"`
}

// MatchGetResponse contains the information stored about a given match
type MatchGetResponse struct {
	ID        int
	UserOne   int
	UserTwo   int
	MatchedOn string
}

// MatchUpdateRequest contains the information required to update a match
type MatchUpdateRequest struct {
	UserOne *int `valid:"-"`
	UserTwo *int `valid:"-"`
}

// MatchListResponse contains the information stored about all matches
type MatchListResponse struct {
	IDs []int
}

// Executes the query, returning the rows
func executeQueryWithResponses(db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(query, args...)
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

// Init opens the database connection
func (dao *DAO) Initialise(config *utils.Config) error {
	connStr := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=%s", config.User, config.DBName, config.Host, config.SSLMode)
	var err error
	dao.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	return nil
}

// CreateMatch inserts a new match into the database given two user IDs
func (dao *DAO) CreateMatch(request MatchCreateRequest) error {
	_, err := executeQuery(dao.DB, "INSERT INTO Match (userOne, userTwo, matchedOn) VALUES ($1, $2, NOW())", request.UserOne, request.UserTwo)
	return err
}

// GetMatch returns the information about a match stored for a given ID
func (dao *DAO) GetMatch(id int64) (*MatchGetResponse, error) {
	row, err := executeQueryWithRowResponse(dao.DB, "SELECT * FROM Match WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	var match MatchGetResponse
	err = row.Scan(&match.ID, &match.UserOne, &match.UserTwo, &match.MatchedOn)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrMatchNotFound(id)
		default:
			return nil, err
		}
	}

	return &match, nil
}

// UpdateMatch updates an already existing match to two user IDs
func (dao *DAO) UpdateMatch(matchID int64, request MatchUpdateRequest) error {
	rowsAffected, err := executeQuery(dao.DB, "UPDATE Match SET userOne = $1, userTwo = $2, matchedOn = NOW() WHERE id = $3", request.UserOne, request.UserTwo, matchID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrMatchNotFound(matchID)
	}

	return nil
}

// DeleteMatch deletes a match from the database
func (dao *DAO) DeleteMatch(id int64) error {
	rowsAffected, err := executeQuery(dao.DB, "DELETE FROM Match WHERE id = $1", id)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrMatchNotFound(id)
	}

	return nil
}

// ListMatch lists all the matches which feature a user
func (dao *DAO) ListMatch(id int64) (*MatchListResponse, error) {
	rows, err := executeQueryWithResponses(dao.DB, "SELECT id FROM Match WHERE userOne = $1 OR userTwo = $1", id)
	if err != nil {
		return nil, err
	}

	var matches MatchListResponse
	matches.IDs = make([]int, 0)
	for rows.Next() {
		var match int
		err = rows.Scan(&match)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				return nil, ErrMatchNotFound(id)
			default:
				return nil, err
			}
		}
		matches.IDs = append(matches.IDs, match)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &matches, nil
}
