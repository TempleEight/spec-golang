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

func (dao *DAO) Init(config *utils.Config) error {
	connStr := fmt.Sprintf("user=%s dbname=%s host=%s sslmode=%s", config.User, config.DBName, config.Host, config.SSLMode)
	var err error
	dao.DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}

	return nil
}