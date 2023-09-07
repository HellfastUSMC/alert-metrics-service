package connectors

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/HellfastUSMC/alert-metrics-service/internal/logger"
	"github.com/HellfastUSMC/alert-metrics-service/internal/server-storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

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
			time.Sleep(time.Second * 1)
			for n := 0; n < 3; n++ {
				err = pg.DBConn.PingContext(ctx)
				if err != nil {
					pg.Logger.Error().Err(err).Msg(fmt.Sprintf("Tried to connect %d times, no luck", n+1))
				} else {
					return nil
				}
				if n != 2 {
					time.Sleep(time.Second * 2)
				}
			}
		}
		pg.Logger.Error().Err(err).Msg("Can't connect to DB, returning")
		return err
	}
	return nil
}

func (pg *PGSQLConn) createTable() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(pg.DBConn, "migrations"); err != nil {
		return err
	}
	return nil
}

// CheckTable checking table exists, creating table if not
func (pg *PGSQLConn) CheckTable() error {
	row := pg.DBConn.QueryRow("SELECT * from Metrics")
	if row.Err() != nil {
		pg.Logger.Info().Msg("Table Metrics not found, trying to create table")
		err := pg.createTable()
		if err != nil {
			pg.Logger.Error().Msg("Can't create table")
			var netErr net.Error
			if errors.As(err, &netErr) {
				pg.Logger.Error().Err(err).Msg("Can't connect to DB server, trying again")
				time.Sleep(time.Second * 1)
				for n := 0; n < 3; n++ {
					err = pg.createTable()
					if err != nil {
						pg.Logger.Error().Err(err).Msg(fmt.Sprintf("Tried to connect %d times, no luck", n+1))
					} else {
						return nil
					}
					if n != 1 {
						time.Sleep(time.Second * 2)
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
func (pg *PGSQLConn) updateMetric(
	metricType string,
	dbTX *sql.Tx,
	ctx context.Context,
	delta serverstorage.Counter,
	val serverstorage.Gauge,
	name string,
) (int64, error) {
	counterUpdateQuery := "UPDATE Metrics SET delta=$1 WHERE name=$2 and type=$3"
	gaugeUpdateQuery := "UPDATE Metrics SET value=$1 WHERE name=$2 and type=$3"
	var rows int64
	if strings.ToUpper(metricType) == GaugeStr {
		pg.Logger.Info().Msg(fmt.Sprintf("Updating %s with value %v of type %s", name, val, metricType))
		res, err := dbTX.ExecContext(ctx, gaugeUpdateQuery, val, name, metricType)
		if err != nil {
			return -1, err
		}
		rows, err = res.RowsAffected()
		if err != nil {
			return -1, err
		}
	} else if strings.ToUpper(metricType) == CounterStr {
		pg.Logger.Info().Msg(fmt.Sprintf("Updating %s with value %v of type %s", name, delta, metricType))
		res, err := dbTX.ExecContext(ctx, counterUpdateQuery, delta, name, metricType)
		if err != nil {
			return -1, err
		}
		rows, err = res.RowsAffected()
		if err != nil {
			return -1, err
		}
	}
	return rows, nil
}

func (pg *PGSQLConn) createMetric(
	metricType string,
	dbTX *sql.Tx,
	ctx context.Context,
	delta serverstorage.Counter,
	val serverstorage.Gauge,
	name string,
) error {
	gaugeInsertQuery := "INSERT INTO Metrics (value,name,type,delta) VALUES ($1,$2,$3,NULL)"
	counterInsertQuery := "INSERT INTO Metrics (delta, name, type, value) VALUES ($1,$2,$3,NULL)"
	var err error
	if strings.ToUpper(metricType) == GaugeStr {
		pg.Logger.Info().Msg(fmt.Sprintf("There's no metric called %s in DB", name))
		pg.Logger.Info().Msg(fmt.Sprintf("Creating %s with value %v of type %s", name, val, metricType))
		_, err = dbTX.ExecContext(ctx, gaugeInsertQuery, val, name, metricType)
	} else if strings.ToUpper(metricType) == CounterStr {
		pg.Logger.Info().Msg(fmt.Sprintf("There's no metric called %s in DB", name))
		pg.Logger.Info().Msg(fmt.Sprintf("Creating %s with value %v of type %s", name, delta, metricType))
		_, err = dbTX.ExecContext(ctx, counterInsertQuery, delta, name, metricType)
	}
	if err != nil {
		pg.Logger.Error().Err(err).Msg("")
		return err
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
		rows, err := pg.updateMetric(GaugeStr, dbTX, ctx, 0, val, name)
		if err != nil {
			pg.Logger.Error().Err(err).Msg("")
		}
		if rows == 0 {
			err = pg.createMetric(GaugeStr, dbTX, ctx, 0, val, name)
			if err != nil {
				return err
			}
		}
	}

	for name, delta := range store.Counter {
		rows, err := pg.updateMetric(CounterStr, dbTX, ctx, delta, 0, name)
		if err != nil {
			pg.Logger.Error().Err(err).Msg("")
		}
		if rows == 0 {
			err = pg.createMetric(CounterStr, dbTX, ctx, delta, 0, name)
			if err != nil {
				return err
			}
		}
	}
	err = dbTX.Commit()
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) {
			pg.Logger.Error().Err(err).Msg("Can't connect to DB server, trying again")
			time.Sleep(time.Second * 1)
			for n := 0; n < 3; n++ {
				err = dbTX.Commit()
				if err != nil {
					pg.Logger.Error().Err(err).Msg(fmt.Sprintf("Tried to connect %d times, but no luck", n+1))
				} else {
					return nil
				}
				if n != 1 {
					time.Sleep(time.Second * 2)
				}
			}
		}
		pg.Logger.Error().Err(err).Msg("Can't commit query to DB, returning")
		err = dbTX.Rollback()
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
			time.Sleep(time.Second * 1)
			for n := 0; n < 3; n++ {
				rows, err = pg.DBConn.QueryContext(ctx, "SELECT * FROM Metrics;")
				if err != nil {
					pg.Logger.Error().Err(err).Msg(fmt.Sprintf("Tried to connect %d times, but no luck", n+1))
				} else {
					err = nil
					break
				}
				if n != 1 {
					time.Sleep(time.Second * 2)
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
