package model2

import (
	"encoding/json"
	"fmt"
	"os"
	//"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/util"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"
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

func EventLogGetPrimaryId() uint64 {

	if !Dbse.BadgerEnable {
		return 0
	}

	for i := range 100 {
		_ = i

		seq, err := Dbse.Badger.GetSequence([]byte("seq_event_log"), 1000)
		if err != nil {
			continue
		}
		defer seq.Release()

		num, err := seq.Next()
		if err != nil {
			continue
		}

		if num == 0 {
			num, err = seq.Next()
			if err != nil {
				continue
			}
		}

		if num > 0 {

			return num
		}
	}

	return 0
}

func EventLogAdd(c *fiber.Ctx, status int, tag, msg string) {

	pref := ""

	if status != 200 {
		pref += red_clr("| ")
	} else {
		pref += green_clr("| ")
	}

	pid := strconv.Itoa(os.Getpid())
	pref += "[" + pid + "] "

	_dt := time.Now().Format("2006-01-02 15:04:05.000")
	pref += "[" + _dt + "] "

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

	event_log_id := uint64(0)
	if Dbse.BadgerEnable && !arg_nolog {
		event_log_id = EventLogGetPrimaryId()
	}

	if Dbse.BadgerEnable && !arg_nolog && event_log_id > 0 {

		msg = util.StringClearColor(msg)

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

		err := Dbse.Badger.Update(func(txn *badger.Txn) error {

			//key := "event_log_" + strconv.FormatUint(event_log_id, 10)
			key := "event_log_" + fmt.Sprintf("%020d", event_log_id)

			err := txn.Set([]byte(key), payload)

			return err
		})
		if err != nil {
			fmt.Println("EventLogAdd Dbse.Badger.Update err=", err.Error())
			panic(err)
		}

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
			prefix := []byte("event_log_")

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				err := item.Value(func(v []byte) error {

					var el EventLog
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {

						EventLogAdd(nil, 500, "EventLogDumpTo", "json.Unmarshal "+err2.Error())
					}

					//fmt.Printf("key=%s, value=%+v\n", k, el)

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

		EventLogAdd(nil, 200, "EventLogDumpTo", "EventLog saved to "+toPath)

	}
}

func EventLogJsonTo(toPath string) {

	if Dbse.BadgerEnable {

		f, err := os.Create(toPath)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		var logs []EventLog

		Dbse.Badger.View(func(txn *badger.Txn) error {

			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("event_log_")

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				err := item.Value(func(v []byte) error {

					var el EventLog
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {

						EventLogAdd(nil, 500, "EventLogJsonTo", "json.Unmarshal "+err2.Error())
					}

					logs = append(logs, el)

					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

		payload, err := json.MarshalIndent(logs, "", "    ")
		if err != nil {
			EventLogAdd(nil, 500, "EventLogJsonTo", "json.MarshalIndent "+err.Error())
		}

		f.WriteString(string(payload))

		EventLogAdd(nil, 200, "EventLogJsonTo", "EventLog saved to "+toPath)

	}
}

func EventLogDump() {

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {

			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("event_log_")

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {

				item := it.Item()
				//k := item.Key()

				err := item.Value(func(v []byte) error {

					var el EventLog
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						//fmt.Println("error:", err2)
						EventLogAdd(nil, 500, "EventLogDump", "json.Unmarshal "+err2.Error())
					}

					//fmt.Printf("key=%s, value=%+v\n", k, el)

					// | [26566] [2026-03-02 08:47:54.070] [192.168.0.110] [login0Xz] [200] [CORE] SendFile /tmp/folder1/morning/mor2/800x800.jpg
					s := fmt.Sprintf("[%d] [%s] [%s] [%s] [%d] [%s] %s", el.ProcId, el.DT, el.IP, el.Login, el.Code, el.Tag, el.Msg)

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

func EventLogTruncate() {

	if Dbse.BadgerEnable {

		prefix := []byte("event_log_")

		err := Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

		// -------------------------------------------------------------------------------------------------------------------------------------------

		prefix = []byte("seq_event_log")

		err = Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

		EventLogAdd(nil, 200, "EventLogTruncate", "EventLog storage recreated")

	}
}
