package fiberinterstate

import (
	//"fmt"
	//"time"
	//"errors"
	//"os"
	//"strconv"
	"context"
	"strings"

	"github.com/gofiber/fiber/v3"
	//ipc "github.com/james-barrow/golang-ipc"
	badger "github.com/dgraph-io/badger/v4"
)

type Config struct {
	App *fiber.App

	BadgerDirectory string

	BadgerHandler *badger.DB

	LogEnable bool
}

// Link to Config
var FiberInterstateConfig Config

var MasterPID int
var Children_pids []int

// Constructor
func New(config ...Config) func(fiber.Ctx) {

	cfg := Config{}
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.App == nil {
		panic("fiberinterstate Config.App is mandatory param")
	}

	if cfg.BadgerDirectory == "" {
		panic("fiberinterstate Config.BadgerDirectory is mandatory param")
	}

	FiberInterstateConfig = cfg

	cfg.App.Hooks().OnListen(func(listenData fiber.ListenData) error {
		if listenData.Prefork {

			MasterPID = listenData.PID
			Children_pids = listenData.ChildPIDs
		}
		return nil
	})

	if !fiber.IsChild() {

		go server_start(cfg)

	} else {

		//go client_start()
	}

	// Return middleware
	return func(c fiber.Ctx) {
		c.Next()
	}
}

// --------------------------------------------------------------------------------------------------------------------------------------

// Start server. Inner function
func server_start(cfg Config) {

	EventLogAdd("IPCServer", "server_start", "starting")

	EventLogAdd("IPCServer", "Start", "connect to badger")
	bh, err := ConnectToBadger(cfg.BadgerDirectory)
	if err != nil {
		panic(err)
	}
	cfg.BadgerHandler = bh

	Dbse.Badger = bh
	Dbse.BadgerEnable = true

	// ----------------------------------------------------------------

	srv := NewIPCServer()

	srv.OnMessage(func(h *hub, c *client, id int, msg string) {

		//fmt.Println("msg=", msg)

		text_command := strings.Split(msg, ":")
		if len(text_command) == 2 {

			//fmt.Println("text_command=", text_command)

			if text_command[0] == "GET" {
				ret, err := BadgerGetOne(text_command[1])

				if err == nil {
					c.outbox <- ret
				} else {

					//fmt.Println("return err=", err.Error(), " for val=", text_command[1])

					c.outbox <- []byte("(empty)")
				}
			}

			if text_command[0] == "DEL" {

				BadgerDel(text_command[1])

				c.outbox <- []byte("(del done)")
			}

		}

		if len(text_command) == 3 {

			//fmt.Println("text_command=", text_command)

			if text_command[0] == "SET" {

				BadgerSet(text_command[1], text_command[2])

				c.outbox <- []byte("(set done)")
			}

		}

	})

	srv.Start()

}

// --------------------------------------------------------------------------------------------------------------------------------------
// Storage signature

/*
func New(config ...Config) Storage
func NewFromConnection(db *badger.DB, config ...Config) *Storage

func (s *Storage) Get(key string) ([]byte, error)
func (s *Storage) GetWithContext(ctx context.Context, key string) ([]byte, error)
func (s *Storage) Set(key string, val []byte, exp time.Duration) error
func (s *Storage) SetWithContext(ctx context.Context, key string, val []byte, exp time.Duration) error
func (s *Storage) Delete(key string) error
func (s *Storage) DeleteWithContext(ctx context.Context, key string) error
func (s *Storage) Reset() error
func (s *Storage) ResetWithContext(ctx context.Context) error

func (s *Storage) Close() error
func (s *Storage) Conn() *badger.DB
*/

// --------------------------------------------------------------------------------------------------------------------------------------

// Get get value by key
func Get(key string) (string, error) {

	if fiber.IsChild() {

		EventLogAdd("FI", "Get", "CHILD", key)

		return ClientImmediateRequest(context.Background(), []byte("GET:"+key))

	} else {

		EventLogAdd("FI", "Get", "MASTER", key)

		bytes_arr, err := BadgerGetOne(key)

		return string(bytes_arr), err
	}
}

// GetWithContext get value by key, aborting if ctx is already done.
func GetWithContext(ctx context.Context, key string) (string, error) {

	if err := ctx.Err(); err != nil {
		return "", err
	}

	return Get(key)
}

// --------------------------------------------------------------------------------------------------------------------------------------

// Set value
func Set(key, val string) error {

	if fiber.IsChild() {

		EventLogAdd("FI", "Set", "CHILD", key)

		_, err := ClientImmediateRequest(context.Background(), []byte("SET:"+key+":"+val))

		return err

	} else {

		EventLogAdd("FI", "Set", "MASTER", key)

		return BadgerSet(key, val)
	}

}

// SetWithContext set value, aborting if ctx is already done.
func SetWithContext(ctx context.Context, key, val string) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	return Set(key, val)

}

// --------------------------------------------------------------------------------------------------------------------------------------

// Delete key
func Delete(key string) error {

	if fiber.IsChild() {

		EventLogAdd("FI", "Delete", "CHILD", key)

		_, err := ClientImmediateRequest(context.Background(), []byte("DEL:"+key))

		return err

	} else {

		EventLogAdd("FI", "Delete", "MASTER", key)

		return BadgerDel(key)
	}

}

// DeleteWithContext delete key, aborting if ctx is already done.
func DeleteWithContext(ctx context.Context, key string) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	return Delete(key)

}

// --------------------------------------------------------------------------------------------------------------------------------------
