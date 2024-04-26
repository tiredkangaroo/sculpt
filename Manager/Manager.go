package Manager

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

var DB *sql.DB
var RequiredConnectionError string = "RequiredConnectionError: There must be connected database to run the %s operation."

func Connected() bool {
	return DB != nil
}
func Connect(user string, password string, db_name string) error {
	database, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", user, password, db_name))
	if err != nil {
		return err
	}
	DB = database
	return nil
}
