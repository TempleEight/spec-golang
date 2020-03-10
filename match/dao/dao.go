package dao

import (
	"database/sql"
	"fmt"

	"github.com/TempleEight/spec-golang/match/util"
	// pq acts as the driver for SQL requests
	_ "github.com/lib/pq"
)

// Datastore provides the interface adopted by the DAO, allowing for mocking
type Datastore interface {
	ListMatch() (*[]Match, error)
	CreateMatch(input CreateMatchInput) (*Match, error)
	ReadMatch(input ReadMatchInput) (*Match, error)
	UpdateMatch(input UpdateMatchInput) (*Match, error)
	DeleteMatch(input DeleteMatchInput) error
}

// DAO encapsulates access to the datastore
type DAO struct {
	DB *sql.DB
}

// Match encapsulates the object stored in the datastore
type Match struct {
	ID        int64
	AuthID    int64
	UserOne   int64
	UserTwo   int64
	MatchedOn string
}

// CreateMatchInput encapsulates the information required to create a single match in the datastore
type CreateMatchInput struct {
	AuthID  int64
	UserOne int64
	UserTwo int64
}

// ReadMatchInput encapsulates the information required to read a single match in the datastore
type ReadMatchInput struct {
	ID int64
}

// UpdateMatchInput encapsulates the information required to update a single match in the datastore
type UpdateMatchInput struct {
	ID      int64
	UserOne int64
	UserTwo int64
}

// DeleteMatchInput encapsulates the information required to delete a single match in the datastore
type DeleteMatchInput struct {
	ID int64
}

// Init opens the datastore connection, returning a DAO
func Init(config *util.Config) (*DAO, error) {
	connStr := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=%s", config.User, config.DBName, config.Host, config.SSLMode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	return &DAO{db}, nil
}

// Executes a query, returning the number of rows affected
func executeQuery(db *sql.DB, query string, args ...interface{}) (int64, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Executes a query, returning the row
func executeQueryWithRowResponse(db *sql.DB, query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...)
}

// Executes a query, returning the rows
func executeQueryWithRowResponses(db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(query, args...)
}

// ListMatch returns a list containing every match in the datastore
func (dao *DAO) ListMatch() (*[]Match, error) {
	rows, err := executeQueryWithRowResponses(dao.DB, "SELECT * FROM Match")
	if err != nil {
		return nil, err
	}

	matchList := make([]Match, 0)
	for rows.Next() {
		var match Match
		err = rows.Scan(&match.ID, &match.AuthID, &match.UserOne, &match.UserTwo, &match.MatchedOn)
		if err != nil {
			return nil, err
		}
		matchList = append(matchList, match)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return &matchList, nil
}

// CreateMatch creates a new match in the datastore, returning the newly created match
func (dao *DAO) CreateMatch(input CreateMatchInput) (*Match, error) {
	row := executeQueryWithRowResponse(dao.DB, "INSERT INTO match (auth_id, userOne, userTwo, matchedOn) VALUES ($1, $2, NOW()) RETURNING *", input.UserOne, input.UserTwo)

	var match Match
	err := row.Scan(&match.ID, &match.AuthID, &match.UserOne, &match.UserTwo, &match.MatchedOn)
	if err != nil {
		return nil, err
	}

	return &match, nil
}

// ReadMatch returns the match in the datastore for a given ID
func (dao *DAO) ReadMatch(input ReadMatchInput) (*Match, error) {
	row := executeQueryWithRowResponse(dao.DB, "SELECT * FROM match WHERE id = $1", input.ID)

	var match Match
	err := row.Scan(&match.ID, &match.AuthID, &match.UserOne, &match.UserTwo, &match.MatchedOn)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrMatchNotFound(input.ID)
		default:
			return nil, err
		}
	}

	return &match, nil
}

// UpdateMatch updates a match in the datastore, returning the newly updated match
func (dao *DAO) UpdateMatch(input UpdateMatchInput) (*Match, error) {
	row := executeQueryWithRowResponse(dao.DB, "UPDATE match SET userOne = $1, userTwo = $2, matchedOn = NOW() WHERE id = $3 RETURNING *", input.UserOne, input.UserTwo, input.ID)

	var match Match
	err := row.Scan(&match.ID, &match.AuthID, &match.UserOne, &match.UserTwo, &match.MatchedOn)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrMatchNotFound(input.ID)
		default:
			return nil, err
		}
	}

	return &match, nil
}

// DeleteMatch deletes a match in the datastore
func (dao *DAO) DeleteMatch(input DeleteMatchInput) error {
	rowsAffected, err := executeQuery(dao.DB, "DELETE FROM match WHERE id = $1", input.ID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrMatchNotFound(input.ID)
	}

	return nil
}
