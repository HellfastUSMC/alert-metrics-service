package connectors

import (
	"database/sql"
	"github.com/HellfastUSMC/alert-metrics-service/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PGSQLConn struct {
	ConnectionString string
	DBConn           *sql.DB
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

func (pg *PGSQLConn) PingBase() error {
	err := pg.DBConn.Ping()
	if err != nil {
		return err
	}
	return nil
}

func NewConnectionPGSQL(config config.SysConfig) (*PGSQLConn, error) {
	db, err := sql.Open("pgx", config.DBPath)
	if err != nil {
		return nil, err
	}
	return &PGSQLConn{
		ConnectionString: config.DBPath,
		DBConn:           db,
	}, nil
}
