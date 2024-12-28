package sql

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var activeDB *pgx.Conn
var Logger = slog.Default()

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func SetActiveDB(db *pgx.Conn) {
	activeDB = db
}

func CloseActiveDB() error {
	return activeDB.Close(context.Background())
}

func Execute(statement string, a ...any) (pgconn.CommandTag, error) {
	Logger.Debug("executing", "statement", statement, "args", a)
	return activeDB.Exec(context.Background(), statement, a...)
}

func Query(statement string, a ...any) (pgx.Rows, error) {
	Logger.Debug("querying", "statement", statement, "args", a)
	return activeDB.Query(context.Background(), statement, a...)
}
