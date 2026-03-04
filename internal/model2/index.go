package model2

import (
	"fmt"
	"os"
	"path"

	"github.com/western/http-here/v2/internal/conf"

	"github.com/gofiber/fiber/v2"

	badger "github.com/dgraph-io/badger/v4"
)

type DBSE struct {
	BadgerEnable bool
	Badger       *badger.DB
}

var Dbse DBSE

func Open() error {

	db1, err := ConnectToBadger()
	if err != nil {
		fmt.Println("ConnectToBadger err=", err)
		panic(err)
	}

	Dbse = DBSE{
		BadgerEnable: true,
		Badger:       db1,
	}

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

				item.Value(func(v []byte) error {

					if string(k) == findKey {

						//fmt.Printf("key=%s, value=%+v\n", k, el)
						//ret = v
						copy(ret, v)
					}

					return nil
				})
			}
			return nil
		})

		if len(ret) > 0 {
			return ret, true
		}
	}

	return ret, false
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
