package sculpt

import (
	"context"
	"sculpt/internals/sql"

	"github.com/jackc/pgx/v5"
)

func Connect(postgresURL string) error {
	conn, err := pgx.Connect(context.Background(), postgresURL)
	if err != nil {
		return err
	}
	sql.SetActiveDB(conn)
	return nil
}

func Close() error {
	return sql.CloseActiveDB()
}
