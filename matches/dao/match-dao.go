package dao

import (
	// pq acts as the driver for SQL requests
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

const connStr = "user=postgres dbname=postgres host=matches-db sslmode=disable"

// MatchGetResponse contains the information stored about a given match
type MatchGetResponse struct {
	ID        int
	userOne   int
	userTwo   int
	matchedOn string
}

// MatchListResponse contains the information stored about all matches
type MatchListResponse struct {
	IDs []int
}

// MatchCreateRequest contains the information required to create a new match
type MatchCreateRequest struct {
	ID1 string `valid:"type(string),required,stringlength(1|255)"`
	ID2 string `valid:"type(string),required,stringlength(1|255)"`
}

// MatchCreateResponse contains the information stored about the newly created match
type MatchCreateResponse struct {
	ID        int
	userOne   int
	userTwo   int
	matchedOn string
}

type MatchUpdateRequest struct {
	ID      int
	userOne int
	userTwo int
}

func executeQueryWithResponses(query string, args ...interface{}) (*sql.Rows, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	rows, err := db.Query(query, args...)
	return rows, err
}

func executeQueryWithRowResponse(query string, args ...interface{}) (*sql.Row, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return db.QueryRow(query, args...), nil
}

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

// GetMatch returns the information about a match stored for a given ID
func GetMatch(id int64) (*MatchGetResponse, error) {
	row, err := executeQueryWithRowResponse("SELECT * FROM Matches WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	var match MatchGetResponse
	err = row.Scan(&match.ID, &match.userOne, &match.userTwo, &match.matchedOn)
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

// ListMatch lists all the matches which feature a user
func ListMatch(id1 int64) (*MatchListResponse, error) {
	rows, err := executeQueryWithResponses("SELECT id FROM Matches WHERE userOne = $1 OR userTwo = $1", id1)
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
				return nil, ErrMatchNotFound(id1)
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

func CreateMatch(request MatchCreateRequest) (*MatchCreateResponse, error) {
	row, err := executeQueryWithRowResponse("INSERT INTO Matches (userOne, userTwo, matchedOn) VALUES ($1, $2, NOW()) RETURNING *", request.ID1, request.ID1)
	if err != nil {
		return nil, err
	}

	var match MatchCreateResponse
	err = row.Scan(&match.ID, &match.userOne, &match.userTwo, &match.matchedOn)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, errors.New("Could not create match")
		default:
			return nil, err
		}
	}

	return &match, nil
}

func UpdateMatch(request MatchUpdateRequest) error {
	rowsAffected, err := executeQuery("UPDATE Matches SET userOne = $1, userTwo = $2 WHERE id = $3", request.userTwo, request.userOne, request.ID)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrMatchNotFound(request.ID)
	}

	return nil
}

func DeleteMatch(id int64) error {
	rowsAffected, err := executeQuery("DELETE FROM Matches WHERE id = $1", id)
	if err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrMatchNotFound(id)
	}

	return nil
}
