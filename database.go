package sculpt

import (
	"context"

	"github.com/tiredkangaroo/sculpt/internals/sql"

	"github.com/jackc/pgx/v5"
)

// Connect connects to the Postgres database using the given URL.
func Connect(postgresURL string) error {
	conn, err := pgx.Connect(context.Background(), postgresURL)
	if err != nil {
		return err
	}
	sql.SetActiveDB(conn)
	return nil
}

// Close closes the active database connection, previously started with Connect.
func Close() error {
	return sql.CloseActiveDB()
}
