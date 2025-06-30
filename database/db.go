package database

import (
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

var DB *sqlx.DB
var err error

func InitDBAndMirate(constr string) {
	DB, err = sqlx.Open("postgres", constr)
	if err != nil {
		logrus.Errorf("failed to connect to postgres: %v", err)
	}
	err = DB.Ping()
	if err != nil {
		logrus.Errorf("failed to ping postgres: %v", err)
	} else {
		logrus.Info("connected to postgres")
	}
	// migration function create to migrate file
	RunningMiration(DB)
}
func RunningMiration(db *sqlx.DB) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		logrus.Errorf("failed to connect to postgres: %v", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file:///home/krishna/Desktop/rms/database/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		logrus.Errorf("failed to running migration: %v", err)
	}
	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logrus.Info("no new migration found")
		} else {
			logrus.Errorf("failed to run migration: %v", err)
		}
	}
}
func CloseDB() {
	err := DB.Close()
	if err != nil {
		logrus.Errorf("failed to close postgres: %v", err)
	}
}

func Tx(db *sqlx.DB, fn func(tx *sqlx.Tx) error) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("start tx failed: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				logrus.Errorf("failed to recover : %v", err)
				return
			}
			panic(p)
		} else if err != nil {
			if rollBackErr := tx.Rollback(); rollBackErr != nil {
				logrus.Errorf("failed to rollback transaction: %v", err)
				return
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				logrus.Errorf("failed to commit transaction: %v", commitErr)
				return
			}
		}
	}()
	err = fn(tx)
	return err
}
