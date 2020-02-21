package utils

type Config struct {
	User    string `json:"user"`
	DBName  string `json:"dbName"`
	Host    string `json:"host"`
	SSLMode string `json:"sslMode"`
}
