package dao

import (
	"database/sql"
	"fmt"

	"github.com/TempleEight/spec-golang/match/util"
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

// MatchUpdateRequest contains the information required to update a match
type MatchUpdateRequest struct {
	UserOne *int `valid:"-"`
	UserTwo *int `valid:"-"`
}

// MatchListResponse contains the information stored about all matches
type MatchListResponse struct {
	MatchList []MatchReadResponse
}

// MatchCreateResponse contains the information about the newly created match
type MatchCreateResponse struct {
	ID        int
	UserOne   int
	UserTwo   int
	MatchedOn string
}

// MatchReadResponse contains the information stored about a given match
type MatchReadResponse struct {
	ID        int
	UserOne   int
	UserTwo   int
	MatchedOn string
}

// MatchUpdateResponse contains the information about the newly updated match
type MatchUpdateResponse struct {
	ID        int
	UserOne   int
	UserTwo   int
	MatchedOn string
}

// Executes the query, returning the rows
func executeQueryWithResponses(db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(query, args...)
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
func (dao *DAO) Init(config *util.Config) error {
	connStr := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=%s", config.User, config.DBName, config.Host, config.SSLMode)
	var err error
	dao.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	return nil
}

// ListMatch returns a list containing every match
func (dao *DAO) ListMatch() (*MatchListResponse, error) {
	rows, err := executeQueryWithResponses(dao.DB, "SELECT * FROM Match")
	if err != nil {
		return nil, err
	}

	var resp MatchListResponse
	resp.MatchList = make([]MatchReadResponse, 0)
	for rows.Next() {
		var match MatchReadResponse
		err = rows.Scan(&match.ID, &match.UserOne, &match.UserTwo, &match.MatchedOn)
		if err != nil {
			return nil, err
		}
		resp.MatchList = append(resp.MatchList, match)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// CreateMatch inserts a new match into the database given two user IDs
func (dao *DAO) CreateMatch(request MatchCreateRequest) (*MatchCreateResponse, error) {
	row := executeQueryWithRowResponse(dao.DB, "INSERT INTO Match (userOne, userTwo, matchedOn) VALUES ($1, $2, NOW()) RETURNING *", request.UserOne, request.UserTwo)

	var resp MatchCreateResponse
	err := row.Scan(&resp.ID, &resp.UserOne, &resp.UserTwo, &resp.MatchedOn)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

// ReadMatch returns the information about a match stored for a given ID
func (dao *DAO) ReadMatch(matchID int64) (*MatchReadResponse, error) {
	row := executeQueryWithRowResponse(dao.DB, "SELECT * FROM Match WHERE id = $1", matchID)

	var resp MatchReadResponse
	err := row.Scan(&resp.ID, &resp.UserOne, &resp.UserTwo, &resp.MatchedOn)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrMatchNotFound(matchID)
		default:
			return nil, err
		}
	}

	return &resp, nil
}

// UpdateMatch updates an already existing match to two user IDs
func (dao *DAO) UpdateMatch(matchID int64, request MatchUpdateRequest) (*MatchUpdateResponse, error) {
	row := executeQueryWithRowResponse(dao.DB, "UPDATE Match SET userOne = $1, userTwo = $2, matchedOn = NOW() WHERE id = $3 RETURNING *", request.UserOne, request.UserTwo, matchID)

	var resp MatchUpdateResponse
	err := row.Scan(&resp.ID, &resp.UserOne, &resp.UserTwo, &resp.MatchedOn)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrMatchNotFound(matchID)
		default:
			return nil, err
		}
	}

	return &resp, nil
}

// DeleteMatch deletes a match from the database
func (dao *DAO) DeleteMatch(matchID int64) error {
	rowsAffected, err := executeQuery(dao.DB, "DELETE FROM Match WHERE id = $1", matchID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrMatchNotFound(matchID)
	}

	return nil
}
