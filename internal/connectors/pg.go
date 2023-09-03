package connectors

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
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

//func (pg *PGSQLConn) Connect() error {
//	db, err := sql.Open("pgx", pg.ConnectionString)
//	if err != nil {
//		return err
//	}
//	pg.DBConn = db
//	return nil
//}

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
		return err
	}
	return nil
}

func (pg *PGSQLConn) WriteDump(jsonString []byte) error {
	//checking table exists, creating table if not
	_, err := pg.DBConn.Query("SELECT * from Metrics LIMIT 1;")
	if err != nil {
		pg.Logger.Info().Msg("Table Metrics not found, creating table")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := pg.DBConn.ExecContext(ctx, `CREATE TABLE Metrics (
			NAME text NOT NULL UNIQUE PRIMARY KEY,
			TYPE text NOT NULL,
			VALUE double precision,
			DELTA bigint;
        )`)
		if err != nil {
			return err
		}
		pg.Logger.Info().Msg("Table Metrics created")
	}
	store := serverstorage.MemStorage{}
	if err := json.Unmarshal(jsonString, &store); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for name, val := range store.Gauge {
		pg.Logger.Info().Msg(fmt.Sprintf("Updating %s with value %f of type Gauge", name, val))
		res, err := pg.DBConn.ExecContext(ctx, "UPDATE Metrics SET value=$1 WHERE name=$2 and type=$3;", val, name, "Gauge")
		rows, _ := res.RowsAffected()
		if err != nil {
			pg.Logger.Error().Err(err)
		}

		if rows == 0 {
			pg.Logger.Info().Msg(fmt.Sprintf("There's no metric called %s in DB", name))
			pg.Logger.Info().Msg(fmt.Sprintf("Creating %s with value %f of type Gauge", name, val))
			_, err := pg.DBConn.ExecContext(ctx, "INSERT INTO Metrics (value,name,type,delta) VALUES ($1,$2,$3,NULL);", val, name, "Gauge")
			if err != nil {
				pg.Logger.Error().Err(err)
			}
		}
	}
	for name, delta := range store.Counter {
		pg.Logger.Info().Msg(fmt.Sprintf("Updating %s with delta %d of type Counter", name, delta))
		//row := pg.DBConn.QueryRow("SELECT delta FROM Metrics WHERE name=$1", name)
		//var d int64
		//if err := row.Scan(&d); err != nil {
		//	pg.Logger.Error().Err(err)
		//}
		res, err := pg.DBConn.ExecContext(ctx, "UPDATE Metrics SET delta=$1 WHERE name=$2 and type=$3;", delta, name, "Counter")
		rows, _ := res.RowsAffected()
		if err != nil {
			pg.Logger.Error().Err(err)
		}
		if rows == 0 {
			pg.Logger.Info().Msg(fmt.Sprintf("There's no metric called %s in DB", name))
			pg.Logger.Info().Msg(fmt.Sprintf("Creating %s with delta %d of type Counter", name, delta))
			_, err := pg.DBConn.ExecContext(ctx, "INSERT INTO Metrics (delta, name, type, value) VALUES ($1,$2,$3,NULL);", delta, name, "Counter")
			if err != nil {
				pg.Logger.Error().Err(err)
			}
		}
	}
	//_, err := pg.DBConn.Prepare("UPDATE Metrics SET ")
	//if err != nil {
	//	return err
	//}
	return nil
}

func (pg *PGSQLConn) ReadDump() ([]string, error) {
	return nil, nil
}

func (pg *PGSQLConn) GetPath() string {
	return pg.ConnectionString
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
