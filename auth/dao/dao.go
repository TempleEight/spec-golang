package dao

import (
	"database/sql"
	"fmt"

	"github.com/TempleEight/spec-golang/auth/utils"
	// pq acts as the driver for SQL requests
	"github.com/lib/pq"
)

// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
const psqlUniqueViolation = "unique_violation"

// Datastore provides the interface adopted by the DAO, allowing for mocking
type Datastore interface {
	CreateAuth(input CreateAuthInput) (*Auth, error)
	ReadAuth(input ReadAuthInput) (*Auth, error)
}

// DAO encapsulates access to the database
type DAO struct {
	DB *sql.DB
}

// Auth encapsulates the object stored in the database
type Auth struct {
	ID       int
	Email    string
	Password string
}

// CreateAuthInput encapsulates the information required to create a single auth
type CreateAuthInput struct {
	Email    string
	Password string
}

// ReadAuthInput encapsulates the information required to read a single auth
type ReadAuthInput struct {
	Email string
}

// Init opens the database connection, returning a DAO
func Init(config *utils.Config) (*DAO, error) {
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

// Executes a query, returning the rows
func executeQueryWithRowResponse(db *sql.DB, query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...)
}

// CreateAuth creates a new auth in the database, returning the newly created auth
func (dao *DAO) CreateAuth(input CreateAuthInput) (*Auth, error) {
	row := executeQueryWithRowResponse(dao.DB, "INSERT INTO auth (email, password) VALUES ($1, $2) RETURNING *", input.Email, input.Password)

	var auth Auth
	err := row.Scan(&auth.ID, &auth.Email, &auth.Password)
	if err != nil {
		// PQ specific error
		if err, ok := err.(*pq.Error); ok {
			if err.Code.Name() == psqlUniqueViolation {
				return nil, ErrDuplicateAuth
			}
		}
		return nil, err
	}

	return &auth, nil
}

// ReadAuth returns the auth for a given email
func (dao *DAO) ReadAuth(input ReadAuthInput) (*Auth, error) {
	row := executeQueryWithRowResponse(dao.DB, "SELECT * FROM auth WHERE email = $1", input.Email)

	var auth Auth
	err := row.Scan(&auth.ID, &auth.Email, &auth.Password)
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
