package sculpt

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var ActiveDB = new(Database)

type Database struct {
	SQLDatabase *sql.DB
}

func (d *Database) Connected() bool {
	return d.SQLDatabase != nil
}

func (d *Database) Execute(statement string) (sql.Result, error) {
	if !ActiveDB.Connected() {
		return nil, OperationRequiresDatabaseConnection("database execution")
	}

	LogInfo("Executing: %s", statement)
	res, err := d.SQLDatabase.Exec(statement)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (d *Database) Query(query string) (*sql.Rows, error) {
	if !ActiveDB.Connected() {
		return nil, OperationRequiresDatabaseConnection("database execution")
	}

	LogInfo("Executing: %s", query)
	res, err := d.SQLDatabase.Query(query)
	if err != nil {
		return nil, err
	}
	return res, err
}

func Connect(user string, password string, db_name string) error {
	database, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", user, password, db_name))
	if err != nil {
		return err
	}
	ActiveDB.SQLDatabase = database
	return nil
}

func Seed(rows ...*Row) {
	for _, r := range rows {
		err := r.Save()
		if err != nil {
			LogError("an error occured during seeding: %s", err.Error())
		}
	}
}
