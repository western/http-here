package model

import (
	"fmt"
	"os"
	"path"
	_ "path/filepath"

	"github.com/western/http-here/internal/conf"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Tabler interface {
	TableName() string
}

func ConnectToSQLite(prefix string) (*gorm.DB, error) {

	if _, err := os.Stat(path.Join(prefix, "db")); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Join(prefix, "db"), os.ModePerm); err != nil {
			return nil, err
		}
	}

    // "?cache=shared&mode=rwc"
	db, err := gorm.Open(sqlite.Open(path.Join(prefix, "db", "registry.db."+conf.Version )), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&EventLog{})
	if err != nil {
		fmt.Println("AutoMigrate EventLog err:", err)
	}

	err = db.AutoMigrate(&File{})
	if err != nil {
		fmt.Println("AutoMigrate File err:", err)
	}

	return db, nil
}
