package sculpt

import (
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/lib/pq"
)

var ActiveDB = new(Database)

type Database struct {
	SQLDatabase *sql.DB
}

func (d *Database) Connected() bool {
	return d.SQLDatabase != nil
}
func (d *Database) Disconnect() {
	d.SQLDatabase = nil
}

func Connect(user string, password string, db_name string) error {
	database, err := sql.Open("postgres", fmt.Sprintf("postgres://%s:%s@localhost/%s?sslmode=disable", user, password, db_name))
	if err != nil {
		return err
	}
	ActiveDB.SQLDatabase = database
	return nil
}

func Connected() bool {
	return ActiveDB.Connected()
}

func Disconnect() {
	ActiveDB.Disconnect()
	return
}

// Seed takes the rows, truncates the models associated with the row,
// and saves the row.
func Seed(rows ...*Row) {
	models := []*Model{}
	for _, r := range rows {
		defer r.Save()
		ok := sliceContains(models, r.Model)
		if !ok {
			models = append(models, r.Model)
		}
	}
	for _, m := range models {
		mCopy := *m
		err := m.DeleteModel()
		if err != nil {
			LogError("an error occured during seeding: %s", err.Error())
			continue
		}
		m := &mCopy
		err = m.Save()
		if err != nil {
			LogError("an error occured during seeding: %s", err.Error())
			continue
		}
	}
}

func (d *Database) Execute(statement string, args ...any) (sql.Result, error) {
	if !ActiveDB.Connected() {
		return nil, OperationRequiresDatabaseConnection("database execution")
	}

	re := regexp.MustCompile(`\$[0-9]+`)
	output := re.ReplaceAllString(statement, "%s")
	LogDebug("Database:", output, args...)

	res, err := d.SQLDatabase.Exec(statement, args...)
	if err != nil {
		return nil, err
	}
	return res, err
}

func (d *Database) Query(query string) (*sql.Rows, error) {
	if !ActiveDB.Connected() {
		return nil, OperationRequiresDatabaseConnection("database execution")
	}

	LogDebug("Database:", "%s", query)
	res, err := d.SQLDatabase.Query(query)
	if err != nil {
		return nil, err
	}
	return res, err
}
