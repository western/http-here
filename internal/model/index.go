package model

import (
	"fmt"
	"os"
	"path"

	"github.com/western/http-here/internal/conf"

	"github.com/gofiber/fiber/v2"

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

	dbPath := path.Join(prefix, "db", "registry.db."+conf.Version)

	dsn := dbPath + "?cache=shared&mode=rwc"

	//dsn := dbPath + "?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_temp_store=MEMORY&_mmap_size=268435456"
	//dsn += "&_busy_timeout=500&_txlock=deferred"

	// -------------------------------------------------------------------------------------------------------------------------------------------

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 logger.Default.LogMode(logger.Silent),
		//PrepareStmt:            true,
	})
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// connection pool

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	/*
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(0) // Connections don't expire
	*/

	sqlDB.SetMaxOpenConns(1)
	//sqlDB.SetConnMaxLifetime(0) // Connections don't expire

	// -------------------------------------------------------------------------------------------------------------------------------------------
	// PRAGMA

	//db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	db.Exec("PRAGMA cache_size=10000")
	db.Exec("PRAGMA temp_store=MEMORY")
	db.Exec("PRAGMA mmap_size=268435456") // 256MB
	db.Exec("PRAGMA optimize")

	// -------------------------------------------------------------------------------------------------------------------------------------------

	err = db.AutoMigrate(&EventLog{})
	if err != nil {
		fmt.Println("AutoMigrate EventLog err:", err)
	}

	err = db.AutoMigrate(&File{})
	if err != nil {
		fmt.Println("AutoMigrate File err:", err)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------
	db.Exec("CREATE INDEX IF NOT EXISTS idx_file_full_path ON file(full_path)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_file_md5 ON file(md5)")

	return db, nil
}

// -------------------------------------------------------------------------------------------------------------------------------------------

func GetBoolFromLocals(c *fiber.Ctx, key string) bool {
	if val, ok := c.Locals(key).(bool); ok {
		return val
	}
	return false
}

func GetStringFromLocals(c *fiber.Ctx, key string, defaultValue string) string {
	if val, ok := c.Locals(key).(string); ok {
		return val
	}
	return defaultValue
}

func GetIntFromLocals(c *fiber.Ctx, key string, defaultValue int) int {
	if val, ok := c.Locals(key).(int); ok {
		return val
	}
	return defaultValue
}
