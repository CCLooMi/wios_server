package conf

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/CCLooMi/sql-mak/mysql"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"time"
)

func initDB(lc fx.Lifecycle, config *Config, log *zap.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Name)
	var db *sql.DB
	var err error
	connect := func() (*sql.DB, error) {
		db, err = sql.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}
		// test connection
		err = db.Ping()
		if err != nil {
			db.Close() // release connection
			return nil, err
		}
		return db, nil
	}
	// try to connect with retries
	var i uint64 = 0
	for {
		db, err = connect()
		if err == nil {
			break
		}
		i++
		log.Warn("failed to connect to database, retrying...",
			zap.Uint64("attempt", i),
			zap.Error(err))
		time.Sleep(1 * time.Second)
	}
	log.Info("database connection established successfully")

	// clean expired sessions
	go sessionCleaner(db, log)

	// ensure that the connection is closed when the lifecycle ends
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing database connection")
			return db.Close()
		},
	})
	return db, nil
}

func sessionCleaner(db *sql.DB, log *zap.Logger) {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	dm := mysql.DELETE().
		FROM("sys_session").
		WHERE("TIMESTAMPDIFF(SECOND,inserted_at,NOW())*1000>expires")
	dm.LOGSQL(false)
	de := dm.Execute(db)
	for range ticker.C {
		r := de.Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Warn("failed to delete expired sessions", zap.Error(err))
		} else if n > 0 {
			log.Info("deleted expired sessions", zap.Int64("deleted", n))
		}
	}
}
