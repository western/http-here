package model2

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"

	"github.com/western/http-here/v2/internal/conf"

	badger "github.com/dgraph-io/badger/v4"
)

var (
	green_clr = color.New(color.FgGreen).SprintFunc()
	red_clr   = color.New(color.FgRed).SprintFunc()
	white_clr = color.New(color.Bold, color.FgWhite).SprintFunc()
)

type EventLog struct {
	ProcId int    `json:"proc_id"`
	DT     string `json:"dt"`

	IP    string `json:"ip"`
	Login string `json:"login"`

	Code int `json:"code"`

	Tag string `json:"tag"`
	Msg string `json:"msg"`
}

// func EventLogAdd(db *badger.DB, c *fiber.Ctx, status, tag, msg string) {
func EventLogAdd(c *fiber.Ctx, status int, tag, msg string) {

	pref := ""

	//status_i, _ := strconv.Atoi(status)

	if status != 200 {
		pref += red_clr("| ")
	} else {
		pref += green_clr("| ")
	}

	pid := strconv.Itoa(os.Getpid())
	pref += "[" + pid + "] "

	_dt := time.Now().Format("2006-01-02 15:04:05.000")
	pref += "[" + _dt + "] "

	_dt_timestamp, _ := time.Parse("2006-01-02 15:04:05.000", _dt)
	//fmt.Println("_dt_timestamp=", _dt_timestamp.Unix())

	_ip := ""
	if c != nil {
		_ip = c.IP()
		pref += "[" + _ip + "] "
	} else {
		pref += "[] "
	}

	arg_silence := false
	arg_nolog := false
	if c != nil {

		arg_silence = GetBoolFromLocals(c, "arg_silence")
		arg_nolog = GetBoolFromLocals(c, "arg_nolog")
	}

	_user := ""
	if c != nil && c.Locals("username") != nil {
		_user = green_clr(c.Locals("username").(string))
	}

	if len(_user) > 0 {
		pref += "[" + _user + "] "
	} else {
		pref += "[] "
	}

	if status != 200 {
		pref += "[" + red_clr(status) + "] "
	} else {
		pref += "[" + green_clr(status) + "] "
	}

	pref += "[" + tag + "] "

	msg = strings.Replace(msg, conf.ArgFold, white_clr(conf.ArgFold), 1)

	pref += msg

	if tag != "INIT" && !arg_silence {
		fmt.Println(pref)
	}

	if Dbse.BadgerEnable && !arg_nolog {

		msg = StringClearColor(msg)

		el := EventLog{
			ProcId: os.Getpid(),
			DT:     _dt,

			IP:    _ip,
			Login: _user,

			Code: status,

			Tag: tag,
			Msg: msg,
		}
		payload, _ := json.Marshal(el)

		//fmt.Println("payload=", string(payload))

		Dbse.Badger.Update(func(txn *badger.Txn) error {

			key := "event_log"

			_dt_timestamp_str := strconv.FormatInt(_dt_timestamp.Unix(), 10)
			key += _dt_timestamp_str

			//fmt.Println("key=", key)

			//err := txn.Set([]byte(key), payload)

			e := badger.NewEntry([]byte(key), payload)
			err := txn.SetEntry(e)

			return err
		})

	}

}

func EventLogDumpTo(toPath string) {

	if Dbse.BadgerEnable {

		f, err := os.Create(toPath)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		Dbse.Badger.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("event_log")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				err := item.Value(func(v []byte) error {

					//fmt.Printf("key=%s, value=%s\n", k, v)

					var el EventLog
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						fmt.Println("error:", err2)
					}

					//fmt.Printf("key=%s, value=%+v\n", k, el)

					/*
					       	EventLog{
					   			ProcId: os.Getpid(),
					   			DT:     _dt,

					   			IP:    _ip,
					   			Login: _user,

					   			Code: status,

					   			Tag: tag,
					   			Msg: msg,
					   		}
					*/

					// | [26566] [2026-03-02 08:47:54.070] [192.168.0.110] [login0Xz] [200] [CORE] SendFile /tmp/folder1/morning/mor2/800x800.jpg
					s := fmt.Sprintf("[%d] [%s] [%s] [%s] [%d] [%s] %s\n", el.ProcId, el.DT, el.IP, el.Login, el.Code, el.Tag, el.Msg)

					f.WriteString(s)

					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

	}
}

func EventLogDump() {

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("event_log")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				err := item.Value(func(v []byte) error {

					//fmt.Printf("key=%s, value=%s\n", k, v)

					var el EventLog
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						fmt.Println("error:", err2)
					}

					//fmt.Printf("key=%s, value=%+v\n", k, el)

					/*
					       	EventLog{
					   			ProcId: os.Getpid(),
					   			DT:     _dt,

					   			IP:    _ip,
					   			Login: _user,

					   			Code: status,

					   			Tag: tag,
					   			Msg: msg,
					   		}
					*/

					// | [26566] [2026-03-02 08:47:54.070] [192.168.0.110] [login0Xz] [200] [CORE] SendFile /tmp/folder1/morning/mor2/800x800.jpg
					s := fmt.Sprintf("[%d] [%s] [%s] [%s] [%d] [%s] %s", el.ProcId, el.DT, el.IP, el.Login, el.Code, el.Tag, el.Msg)

					//f.WriteString(s)
					fmt.Println(s)

					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

	}
}

func EventLogClear() {

	if Dbse.BadgerEnable {

		prefix := []byte("event_log")

		err := Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

	}
}

func StringClearColor(msg string) string {

	// /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, ''

	re := regexp.MustCompile("[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]")
	return re.ReplaceAllString(msg, "")
}
