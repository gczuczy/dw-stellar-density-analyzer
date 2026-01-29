package db

import (
	"fmt"
	"time"
	"errors"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	//"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5"

	"github.com/gczuczy/dw-stellar-density-analyzer/pkg/config"
	ds "github.com/gczuczy/dw-stellar-density-analyzer/pkg/densitysurvey"
)

var (
	Pool *DBPool=nil

	prepared = map[string]string{
		"addsheetmeasurement": `
SELECT density.addsheetmeasurement($1::text, $2::text)
`,
		// measurementid, sysname, x,y,z, syscount, maxdistance
		"adddatapoint": `
INSERT INTO density.datapoints (measurementid, sysname, zsample, x,y,z, syscount, maxdistance)
VALUES ($1::int, $2::text, $3::int, $4::real, $5::real, $6::real, $7::int, $8::real)
`,
	}
)

type DBPool struct {
	ctx context.Context
	pool *pgxpool.Pool
}

// init the DBPool and store it in the global variable
func Init(cfg *config.DBConfig) error {
	var err error

	dbcfg, err := pgxpool.ParseConfig("")
	if err != nil {
		return err
	}
	dbcfg.MaxConnLifetime = 8 * time.Hour
	dbcfg.MaxConns = cfg.MaxConns
	dbcfg.MinConns = cfg.MinConns
	dbcfg.AfterConnect = afterConn
	dbcfg.ConnConfig.Host = cfg.Host
	dbcfg.ConnConfig.Port = 5432
	dbcfg.ConnConfig.Database = cfg.Database
	dbcfg.ConnConfig.User = cfg.User
	dbcfg.ConnConfig.Password = cfg.Password
	if cfg.Port != nil {
		dbcfg.ConnConfig.Port = (*cfg.Port)
	}

	dbp := DBPool{
		ctx: context.Background(),
	}

	if dbp.pool, err = pgxpool.NewWithConfig(dbp.ctx, dbcfg); err != nil {
		return err
	}

	conn, err := dbp.pool.Acquire(dbp.ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	Pool = &dbp
	return nil
}

func afterConn(ctx context.Context, dbc *pgx.Conn) error {
	for name, query := range prepared {
		if _, err := dbc.Prepare(ctx, name, query); err != nil {
			return errors.Join(err, fmt.Errorf("Error while preparing %s", name))
		}
	}
	return nil
}

func (p *DBPool) AddMeasurement(m *ds.Measurement) (err error) {
	conn, err := p.pool.Acquire(p.ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	tx, err := conn.Begin(p.ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(p.ctx)
			return
		}
		if tx.Commit(p.ctx) != nil {
			tx.Rollback(p.ctx)
		}
	}()

	var rows pgx.Rows

	if rows, err = tx.Query(p.ctx, "addsheetmeasurement",	m.CMDR, m.Project);  err != nil {
		return err
	}

	if !rows.Next() {
		return fmt.Errorf("No measurementid returned")
	}

	var vs []any
	if vs, err = rows.Values(); err != nil {
		return errors.Join(err, fmt.Errorf("Fuck golang's error handling"))
	}

	mid, ok := vs[0].(int32)
	if !ok {
		rows.Close()
		return errors.Join(err, fmt.Errorf("Fuck golang's error handling again, %v/%T -> %v", vs[0], vs[0], mid))
	}
	rows.Close()

	for _, dp := range m.DataPoints {
		if _, err = tx.Exec(p.ctx, "adddatapoint", mid, dp.SystemName, dp.ZSample,
			dp.X, dp.Y, dp.Z, dp.Count, dp.MaxDistance); err != nil {
			return errors.Join(err, fmt.Errorf("Error while inserting datapoint"))
		}
	}

	return nil
}
