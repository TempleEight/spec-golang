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

// AuthCreateRequest contains the information required to create a new authorized user
type AuthCreateRequest struct {
	Email    string `valid:"email,required"`
	Password string `valid:"type(string),required,stringlength(8|64)"`
}

type AuthCreateResponse struct {
	AccessToken string
}

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

func (dao *DAO) CreateAuth(request AuthCreateRequest) error {
	_, err := executeQuery(dao.DB, "INSERT INTO auth (email, password) VALUES ($1, $2)", request.Email, request.Password)
	return err
}
