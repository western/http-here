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

func ConnectToSQLite() (*gorm.DB, error) {

	homepath, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		/*
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code": 500,
				"msg":  "Error homedir detect",
			}, "application/json")
		*/
	}

	if _, err := os.Stat(path.Join(homepath, ".httphere", "db")); err != nil {

		if err := os.MkdirAll(path.Join(homepath, ".httphere", "db"), os.ModePerm); err != nil {
			fmt.Println(err)
			/*
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code": 500,
					"msg":  "Error db folder create",
				}, "application/json")
			*/
		}
	}

	db, err := gorm.Open(sqlite.Open(path.Join(homepath, ".httphere", "db", "registry.db."+conf.Version)), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&EventLog{})
	if err != nil {
		fmt.Println(err)
	}

	err = db.AutoMigrate(&File{})
	if err != nil {
		fmt.Println(err)
	}

	return db, nil
}
