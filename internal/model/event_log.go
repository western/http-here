package model

import (
	"fmt"
	"os"
	_ "reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gofiber/fiber/v2"

	"gorm.io/gorm"
)

type EventLog struct {
	Id uint `json:"id" gorm:"unique;primaryKey;autoIncrement"`

	Proc_id int
	DT      string

	IP    string
	Login string

	Code string

	TAG string
	MSG string
}

func (EventLog) TableName() string {
	return "event_log"
}

var (
	green_clr = color.New(color.FgGreen).SprintFunc()
	red_clr   = color.New(color.FgRed).SprintFunc()
	white_clr = color.New(color.Bold, color.FgWhite).SprintFunc()
)

// model.EventLogAdd(db, c, "500", "CORE", "Error "+err.Error())
// model.EventLogAdd(nil, nil, "500", "CORE", "Error "+err.Error())
func EventLogAdd(db *gorm.DB, c *fiber.Ctx, status, tag, msg string) {

	pref := ""

	if status != "200" {
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

	arg_fold := ""
	arg_silence := ""
	arg_nolog := ""
	if c != nil {

		arg_fold = c.Locals("arg_fold").(string)
		if c.Locals("arg_silence") != nil {
			arg_silence = c.Locals("arg_silence").(string)
		}
		if c.Locals("arg_nolog") != nil {
			arg_nolog = c.Locals("arg_nolog").(string)
		}
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

	if status != "200" {
		pref += "[" + red_clr(status) + "] "
	} else {
		pref += "[" + green_clr(status) + "] "
	}

	pref += "[" + tag + "] "

	msg = strings.Replace(msg, arg_fold, white_clr(arg_fold), 1)

	pref += msg

	if tag != "INIT" && arg_silence == "" {
		fmt.Println(pref)
	}

	if db != nil && arg_nolog == "" {

		msg = StringClearColor(msg)

		el := &EventLog{
			Proc_id: os.Getpid(),
			DT:      _dt,

			IP:    _ip,
			Login: _user,

			Code: status,

			TAG: tag,
			MSG: msg,
		}

		result := db.Create(el)
		if result.Error != nil {
			//return result.Error
			fmt.Println(result.Error)
		}
	}

}

func StringClearColor(msg string) string {

	// /[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]/g, ''

	re := regexp.MustCompile("[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]")
	return re.ReplaceAllString(msg, "")
}

/*
func EventLogAdd(db *gorm.DB, tag, msg string) error {

	pid := os.Getpid()

	dt := time.Now().Format("2006-01-02 15:04:05.000")

	el := &EventLog{
		Proc_id: pid,
		DT:      dt,

		IP:    "",
		Login: "",

		Code: "",

		TAG: tag,
		MSG: msg,
	}

	result := db.Create(el)
	if result.Error != nil {
		return result.Error
	}
	return nil

}

func EventLogMsg(db *gorm.DB, c *fiber.Ctx, code, tag, msg string) error {

	pid := os.Getpid()

	dt := time.Now().Format("2006-01-02 15:04:05.000")

	username := ""
	if c.Locals("username") != nil {
		username = c.Locals("username").(string)
	}

	el := &EventLog{
		Proc_id: pid,
		DT:      dt,

		IP:    c.IP(),
		Login: username,

		Code: code,

		TAG: tag,
		MSG: msg,
	}

	result := db.Create(el)
	if result.Error != nil {
		return result.Error
	}
	return nil

}
*/
