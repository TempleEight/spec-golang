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

// AuthCreateRequest contains the information retrieved from an end user to create a new auth
type AuthCreateRequest struct {
	Email    string `valid:"email,required"`
	Password string `valid:"type(string),required,stringlength(8|64)"`
}

// AuthReadRequest contains the information retrieved from an end user to validate an existing auth
type AuthReadRequest struct {
	Email    string `valid:"email,required"`
	Password string `valid:"type(string),required,stringlength(8|64)"`
}

// AuthCreateResponse contains an access token associated to a given auth
type AuthCreateResponse struct {
	AccessToken string
}

// AuthReadResponse contains an access token associated to a given auth
type AuthReadResponse struct {
	AccessToken string
}

// Auth contains the full information persisted in the datastore
type Auth struct {
	Id       string
	Email    string
	Password string
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
func (dao *DAO) CreateAuth(request AuthCreateRequest) (*Auth, error) {
	row, err := executeQueryWithRowResponse(dao.DB, "INSERT INTO auth (email, password) VALUES ($1, $2) RETURNING *", request.Email, request.Password)
	var auth Auth
	err = row.Scan(&auth.Id, &auth.Email, &auth.Password)
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

// ReadAuth attempts to find an existing auth'd user in the data store
func (dao *DAO) ReadAuth(request AuthReadRequest) (*Auth, error) {
	row, err := executeQueryWithRowResponse(dao.DB, "SELECT * FROM auth WHERE email = $1", request.Email)
	var auth Auth
	err = row.Scan(&auth.Id, &auth.Email, &auth.Password)
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
