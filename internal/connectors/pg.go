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

// PGSQLConn Структура взаимодействия с БД
type PGSQLConn struct {
	Logger           logger.CLogger
	DBConn           *sql.DB
	ConnectionString string
}

const (
	GaugeStr   = "GAUGE"
	CounterStr = "COUNTER"
)

// Close Функция закрытия соединения с БД
func (pg *PGSQLConn) Close() error {
	err := pg.DBConn.Close()
	if err != nil {
		return err
	}
	return nil
}

// Ping Функция проверки соединения с БД
func (pg *PGSQLConn) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	f := func() error {
		if err := pg.DBConn.PingContext(ctx); err != nil {
			return err
		}
		return nil
	}
	var netErr net.Error
	err := retryWriteFunc(2, 3, f, &netErr)
	if err != nil {
		return err
	}
	return nil
}

func retryReadFunc(
	interval int,
	attempts int,
	readFunc func() (*sql.Rows, error),
	errorToRetry *net.Error,
) (*sql.Rows, error) {
	if errorToRetry == nil {
		return nil, fmt.Errorf("please provide error to retry to")
	}
	if readFunc == nil {
		return nil, fmt.Errorf("no read func provided")
	}
	rows, err := readFunc()
	if err != nil {
		if errors.As(err, errorToRetry) {
			for i := 0; i < attempts; i++ {
				time.Sleep(time.Second * time.Duration(interval))
				rows, err = readFunc()
				if err == nil {
					return rows, nil
				}
			}
		}
		return nil, err
	}
	return rows, nil
}

func retryWriteFunc(
	interval int,
	attempts int,
	writeFunc func() error,
	errorToRetry *net.Error,
) error {
	if errorToRetry == nil {
		return fmt.Errorf("please provide error to retry to")
	}
	if writeFunc == nil {
		return fmt.Errorf("no write func provided")
	}
	err := writeFunc()
	if err != nil {
		if errors.As(err, errorToRetry) {
			for i := 0; i < attempts; i++ {
				time.Sleep(time.Second * time.Duration(interval))
				err = writeFunc()
				if err == nil {
					return nil
				}
			}
		}
		return err
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

// WriteDump Функция записи дампа метрик в БД
func (pg *PGSQLConn) WriteDump(jsonString []byte) error {
	if err := pg.Ping(); err != nil {
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
		rows, err1 := pg.updateMetric(GaugeStr, dbTX, ctx, 0, val, name)
		if err1 != nil {
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
		rows, err2 := pg.updateMetric(CounterStr, dbTX, ctx, delta, 0, name)
		if err2 != nil {
			pg.Logger.Error().Err(err).Msg("")
		}
		if rows == 0 {
			err = pg.createMetric(CounterStr, dbTX, ctx, delta, 0, name)
			if err != nil {
				return err
			}
		}
	}
	f := func() error {
		if err3 := dbTX.Commit(); err != nil {
			return err3
		}
		return nil
	}
	var netErr net.Error
	err = retryWriteFunc(2, 3, f, &netErr)
	if err != nil {
		return err
	}
	pg.Logger.Info().Msg("Updated")
	return nil
}

// ReadDump Функция чтения дампа метрик из БД
func (pg *PGSQLConn) ReadDump() ([]string, error) {
	pg.Logger.Info().Msg("Reading dump from DB")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	var netErr net.Error
	f := func() (*sql.Rows, error) {
		rows, err := pg.DBConn.QueryContext(ctx, "SELECT * FROM Metrics;")
		if err != nil {
			return nil, err
		}
		return rows, nil
	}
	rows, err := retryReadFunc(2, 3, f, &netErr)
	if err != nil {
		return nil, err
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
	mStore := serverstorage.MemStorage{
		Gauge:   map[string]serverstorage.Gauge{},
		Counter: map[string]serverstorage.Counter{},
	}
	rCount := 0
	for rows.Next() {
		err := rows.Scan(&name, &mType, &value, &delta)
		if err != nil {
			pg.Logger.Error().Err(err).Msg("")
			return nil, err
		}
		if strings.ToUpper(mType) == GaugeStr {
			mStore.Gauge[name] = serverstorage.Gauge(value.Float64)
		}
		if strings.ToUpper(mType) == CounterStr {
			mStore.Counter[name] = serverstorage.Counter(delta.Int64)
		}
		jsonStore, err := json.Marshal(mStore)
		if err != nil {
			return nil, err
		}
		res = append(res, string(jsonStore))
		res = append(res, "\n")
		rCount++
	}
	if rCount == 0 {
		return nil, fmt.Errorf("nothing to read from table")
	}
	return res, nil
}

// NewConnectionPGSQL Функция инициализации новой структуры взаимодействия с БД
func NewConnectionPGSQL(connPath string, logger logger.CLogger) (*PGSQLConn, error) {
	db, err := sql.Open("pgx", connPath)
	if err != nil {
		return nil, err
	}
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}
	return &PGSQLConn{
		ConnectionString: connPath,
		DBConn:           db,
		Logger:           logger,
	}, nil
}
