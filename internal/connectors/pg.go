package connectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PGSQLConn struct {
	ConnectionString string
	DBConn           *sql.DB
	Logger           logger.CLogger
}

const (
	GaugeStr   = "GAUGE"
	CounterStr = "COUNTER"
)

func (pg *PGSQLConn) Close() error {
	err := pg.DBConn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (pg *PGSQLConn) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	err := pg.DBConn.PingContext(ctx)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			pg.Logger.Error().Err(err).Msg("Can't connect to DB server, trying again")
			for n := 1; n <= 5; n = n + 2 {
				time.Sleep(time.Second * time.Duration(n))
				err = pg.DBConn.PingContext(ctx)
				if err != nil {
					pg.Logger.Error().Err(err).Msg(fmt.Sprintf("Tried connect again after %ds", n))
				} else {
					return nil
				}
			}
		}
		pg.Logger.Error().Err(err).Msg("Can't connect to DB, returning")
		return err
	}
	return nil
}

// CheckTable checking table exists, creating table if not
func (pg *PGSQLConn) CheckTable() error {
	row := pg.DBConn.QueryRow("SELECT * from Metrics")
	if row.Err() != nil {
		pg.Logger.Info().Msg("Table Metrics not found, trying to create table")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := pg.DBConn.ExecContext(ctx, `CREATE TABLE Metrics (
			NAME text NOT NULL UNIQUE PRIMARY KEY,
			TYPE text NOT NULL,
			VALUE double precision,
			DELTA bigint
        )`)
		if err != nil {
			pg.Logger.Error().Msg("Can't create table")
			var netErr net.Error
			if errors.As(err, &netErr) {
				pg.Logger.Error().Err(err).Msg("Can't connect to DB server, trying again")
				for n := 1; n <= 5; n = n + 2 {
					time.Sleep(time.Second * time.Duration(n))
					_, err := pg.DBConn.ExecContext(ctx, `CREATE TABLE Metrics (
						NAME text NOT NULL UNIQUE PRIMARY KEY,
						TYPE text NOT NULL,
						VALUE double precision,
						DELTA bigint
					)`)
					if err != nil {
						pg.Logger.Error().Err(err).Msg(fmt.Sprintf("Tried to connect after %ds, but no luck", n))
					} else {
						return nil
					}
				}
			}
			pg.Logger.Error().Err(err).Msg("Can't connect to DB, returning")
			return err
		}
		pg.Logger.Info().Msg("Table Metrics created")
	}
	return nil
}

func (pg *PGSQLConn) WriteDump(jsonString []byte) error {
	if err := pg.CheckTable(); err != nil {
		return err
	}
	store := serverstorage.MemStorage{}
	if err := json.Unmarshal(jsonString, &store); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	dbTX, err := pg.DBConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	for name, val := range store.Gauge {
		pg.Logger.Info().Msg(fmt.Sprintf("Updating %s with value %f of type Gauge", name, val))
		res, err := dbTX.ExecContext(ctx, "UPDATE Metrics SET value=$1 WHERE name=$2 and type=$3", val, name, "Gauge")
		row, _ := res.RowsAffected()
		if err != nil {
			pg.Logger.Error().Err(err).Msg("")
		}

		if row == 0 {
			pg.Logger.Info().Msg(fmt.Sprintf("There's no metric called %s in DB", name))
			pg.Logger.Info().Msg(fmt.Sprintf("Creating %s with value %f of type Gauge", name, val))
			_, err := dbTX.ExecContext(ctx, "INSERT INTO Metrics (value,name,type,delta) VALUES ($1,$2,$3,NULL)", val, name, "Gauge")
			if err != nil {
				pg.Logger.Error().Err(err).Msg("")
			}
		}
	}
	for name, delta := range store.Counter {
		pg.Logger.Info().Msg(fmt.Sprintf("Updating %s with delta %d of type Counter", name, delta))
		res, err := dbTX.ExecContext(ctx, "UPDATE Metrics SET delta=$1 WHERE name=$2 and type=$3", delta, name, "Counter")
		rows, _ := res.RowsAffected()
		if err != nil {
			pg.Logger.Error().Err(err).Msg("")
		}
		if rows == 0 {
			pg.Logger.Info().Msg(fmt.Sprintf("There's no metric called %s in DB", name))
			pg.Logger.Info().Msg(fmt.Sprintf("Creating %s with delta %d of type Counter", name, delta))
			_, err := dbTX.ExecContext(ctx, "INSERT INTO Metrics (delta, name, type, value) VALUES ($1,$2,$3,NULL)", delta, name, "Counter")
			if err != nil {
				pg.Logger.Error().Err(err).Msg("")
			}
		}
	}
	err = dbTX.Commit()
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			pg.Logger.Error().Err(err).Msg("Can't connect to DB server, trying again")
			for n := 1; n <= 5; n = n + 2 {
				time.Sleep(time.Second * time.Duration(n))
				err = dbTX.Commit()
				if err != nil {
					pg.Logger.Error().Err(err).Msg(fmt.Sprintf("Tried to connect after %ds, but no luck", n))
				} else {
					return nil
				}
			}
		}
		pg.Logger.Error().Err(err).Msg("Can't commit query to DB, returning")
		return err
	}
	return nil
}

func (pg *PGSQLConn) ReadDump() ([]string, error) {
	pg.Logger.Info().Msg("Reading dump from DB")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	rows, err := pg.DBConn.QueryContext(ctx, "SELECT * FROM Metrics;")
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			pg.Logger.Error().Err(err).Msg("Can't connect to DB server, trying again")
			for n := 1; n <= 5; n = n + 2 {
				time.Sleep(time.Second * time.Duration(n))
				rows, err = pg.DBConn.QueryContext(ctx, "SELECT * FROM Metrics;")
				if err != nil {
					pg.Logger.Error().Err(err).Msg(fmt.Sprintf("Tried to connect after %ds, but no luck", n))
				} else {
					err = nil
					break
				}
			}
		}
		if err != nil {
			return nil, err
		}
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	var (
		name  string
		mType string
		delta sql.NullInt64
		value sql.NullFloat64
	)
	res := []string{}
	jsonGStr := `{"Gauge":{`
	jsonCStr := `"Counter":{`
	for rows.Next() {
		err := rows.Scan(&name, &mType, &value, &delta)
		if err != nil {
			pg.Logger.Error().Err(err).Msg("")
			return nil, err
		}
		if strings.ToUpper(mType) == GaugeStr {
			val, _ := value.Value()
			jsonGStr += fmt.Sprintf(`"%s":%f,`, name, val)
		} else if strings.ToUpper(mType) == CounterStr {
			del, _ := delta.Value()
			jsonCStr += fmt.Sprintf(`"%s":%d,`, name, del)
		}
	}
	jsonGStr = strings.TrimSuffix(jsonGStr, ",")
	jsonCStr = strings.TrimSuffix(jsonCStr, ",")
	jsonGStr += "},"
	jsonCStr += "}"
	resString := jsonGStr + jsonCStr + "}"
	res = append(res, resString)
	res = append(res, "\n")
	return res, nil
}

func (pg *PGSQLConn) GetPath() string {
	return "Metrics table from DB"
}

func NewConnectionPGSQL(connPath string, logger logger.CLogger) (*PGSQLConn, error) {
	db, err := sql.Open("pgx", connPath)
	if err != nil {
		return nil, err
	}
	return &PGSQLConn{
		ConnectionString: connPath,
		DBConn:           db,
		Logger:           logger,
	}, nil
}
