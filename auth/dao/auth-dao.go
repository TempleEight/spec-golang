package dao

import (
	"database/sql"
	"fmt"

	"github.com/TempleEight/spec-golang/auth/utils"
	// pq acts as the driver for SQL requests
	_ "github.com/lib/pq"
)

// DAO encapsulates access to the database
type DAO struct {
	DB *sql.DB
}

// CreateAuthRequest contains the information retrieved from an end user to create a new auth
type CreateAuthRequest struct {
	Email    string `valid:"email,required"`
	Password string `valid:"type(string),required,stringlength(8|64)"`
}

// ReadAuthRequest contains the information retrieved from an end user to validate an existing auth
type ReadAuthRequest struct {
	Email    string `valid:"email,required"`
	Password string `valid:"type(string),required,stringlength(8|64)"`
}

// CreateAuthResponse contains an access token associated to a given auth
type CreateAuthResponse struct {
	AccessToken string
}

// ReadAuthResponse contains an access token associated to a given auth
type ReadAuthResponse struct {
	AccessToken string
}

// Init constructs a DAO from a configuration file
func (dao *DAO) Init(config *utils.Config) error {
	connStr := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=%s", config.User, config.DBName, config.Host, config.SSLMode)
	var err error
	dao.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	return nil
}

// Executes a query, returning the number of rows affected
func executeQuery(db *sql.DB, query string, args ...interface{}) (int64, error) {
	result, err := db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func executeQueryWithRowResponse(db *sql.DB, query string, args ...interface{}) (*sql.Row, error) {
	return db.QueryRow(query, args...), nil
}

// CreateAuth persists a new auth'd user to the data store
func (dao *DAO) CreateAuth(request CreateAuthRequest) error {
	_, err := executeQuery(dao.DB, "INSERT INTO auth (email, password) VALUES ($1, $2)", request.Email, request.Password)
	return err
}

// GetAuth attempts to find an existing auth'd user in the data store
func (dao *DAO) ReadAuth(request ReadAuthRequest) (*ReadAuthRequest, error) {
	row, err := executeQueryWithRowResponse(dao.DB, "SELECT email, password FROM auth WHERE email = $1", request.Email)
	var auth ReadAuthRequest
	err = row.Scan(&auth.Email, &auth.Password)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, ErrAuthNotFound
		default:
			return nil, err
		}
	}

	return &auth, nil
}
