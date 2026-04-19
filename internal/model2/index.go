package model2

import (
	"fmt"
	"os"
	"path"

	"github.com/western/http-here/v2/internal/conf"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/gofiber/fiber/v2"
	//"github.com/valyala/fasthttp"
)

type DBSE struct {
	BadgerEnable bool
	Badger       *badger.DB

	LogToEnable bool
	LogToPath   string
	LogTo       *os.File

	JsonToEnable bool
	JsonToPath   string
	JsonTo       *os.File
}

var Dbse DBSE

func Open() error {

	db1, err := ConnectToBadger()
	if err != nil {
		fmt.Println("ConnectToBadger err=", err)
		panic(err)
	}

	Dbse.BadgerEnable = true
	Dbse.Badger = db1

	return nil
}

func LogtoOpen(fileName string) error {

	fh, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
		return err
	}

	Dbse.LogToEnable = true
	Dbse.LogToPath = fileName
	Dbse.LogTo = fh

	return nil
}

func JsontoOpen(fileName string) error {

	fh, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
		return err
	}

	Dbse.JsonToEnable = true
	Dbse.JsonToPath = fileName
	Dbse.JsonTo = fh

	//Dbse.JsonTo.WriteString("[\n")

	return nil
}

// -------------------------------------------------------------------------------------------------------------------------------------------

func ConnectToBadger() (*badger.DB, error) {

	if _, err := os.Stat(path.Join(conf.ConfigRoot, "db")); os.IsNotExist(err) {
		if err := os.MkdirAll(path.Join(conf.ConfigRoot, "db"), os.ModePerm); err != nil {
			return nil, err
		}
	}

	dbPath := path.Join(conf.ConfigRoot, "db", "registry.bd."+conf.Version)

	// -------------------------------------------------------------------------------------------------------------------------------------------

	opts := badger.DefaultOptions(dbPath).
		WithLoggingLevel(badger.ERROR)

	//db, err := badger.Open(badger.DefaultOptions(dbPath))
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	return db, nil

}

func BadgerGetOne(findKey string) ([]byte, bool) {

	if len(findKey) == 0 {
		panic("Yous should set findKey")
	}

	var ret []byte

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {

			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte(findKey)

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {

				item := it.Item()
				k := item.Key()

				if string(k) == findKey {

					item.Value(func(v []byte) error {

						ret = v
						return nil
					})
				}

			}
			return nil
		})

		if len(ret) > 0 {
			return ret, true
		}
	}

	return ret, false
}

func BadgerDumpTo(toFile string) {

	if Dbse.BadgerEnable {

		var err error
		var srcfd *os.File

		srcfd, err = os.OpenFile(toFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			panic(err)
		}
		defer srcfd.Close()

		_, err = Dbse.Badger.Backup(srcfd, 0)
		if err != nil {
			panic(err)
		}

		EventLogAdd(nil, 200, "BadgerDumpTo", "Database dumped to "+toFile)
	}
}

func BadgerRestoreFrom(fromFile string) {

	if Dbse.BadgerEnable {

		var err error
		var srcfd *os.File

		srcfd, err = os.Open(fromFile)
		if err != nil {
			panic(err)
		}
		defer srcfd.Close()

		err = Dbse.Badger.Load(srcfd, 16)
		if err != nil {
			panic(err)
		}

		EventLogAdd(nil, 200, "BadgerRestoreFrom", "Database restored from "+fromFile)
	}
}

func BadgerDestroy() {

	if Dbse.BadgerEnable {

		err := Dbse.Badger.DropAll()
		if err != nil {
			panic(err)
		}

		EventLogAdd(nil, 200, "BadgerDestroy", "Database destroyed")
	}
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
