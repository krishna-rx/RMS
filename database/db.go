package database

import (
	"errors"
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
