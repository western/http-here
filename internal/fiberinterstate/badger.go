package fiberinterstate

import (
	//"fmt"
	"os"
	//"path"
	"errors"

	badger "github.com/dgraph-io/badger/v4"
	//"github.com/gofiber/fiber/v3"
)

type DBSE struct {
	BadgerEnable bool
	Badger       *badger.DB
}

var Dbse DBSE

func ConnectToBadger(pathToBadger string) (*badger.DB, error) {

	if _, err := os.Stat(pathToBadger); err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	opts := badger.DefaultOptions(pathToBadger)

	// one of DEBUG, INFO, WARNING or ERROR levels
	// The default value of LoggingLevel is INFO

	//opts = opts.WithLoggingLevel(badger.INFO)
	opts = opts.WithLoggingLevel(badger.ERROR)

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	return db, nil

}

func BadgerGetOne(findKey string) ([]byte, error) {

	if !Dbse.BadgerEnable {
		return nil, errors.New("BadgerSet: Badger is offline")
	}

	if len(findKey) == 0 {

		return nil, errors.New("BadgerSet: findKey is a mandatory param")
	}

	var ret []byte

	err := Dbse.Badger.View(func(txn *badger.Txn) error {

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

	if err != nil {
		return nil, err
	}

	if len(ret) > 0 {
		return ret, nil
	}

	return nil, nil
}

func BadgerSet(key, payload string) error {

	if !Dbse.BadgerEnable {
		return errors.New("BadgerSet: Badger is offline")
	}

	if len(key) == 0 {
		return errors.New("BadgerSet: key is a mandatory param")
	}

	err := Dbse.Badger.Update(func(txn *badger.Txn) error {

		err := txn.Set([]byte(key), []byte(payload))
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func BadgerDel(key string) error {

	if !Dbse.BadgerEnable {
		return errors.New("BadgerDel: Badger is offline")
	}

	if len(key) == 0 {
		return errors.New("BadgerDel: key is a mandatory param")
	}

	prefix := []byte(string(key))

	err := Dbse.Badger.DropPrefix(prefix)
	if err != nil {
		return err
	}

	return nil
}
