package model

import (
	"os"
	_ "strconv"
	"time"

	_ "github.com/fatih/color"
	"github.com/gofiber/fiber/v2"

	_ "gorm.io/driver/sqlite"
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

func EventLogAdd(db *gorm.DB, tag, msg string) error {

	/*
	   db, err := connectToSQLite()
	   if err != nil {
	       log.Fatal(err)
	   }
	   defer db.Close()
	*/

	pid := os.Getpid()
	//pid := strconv.Itoa(os.Getpid())

	dt := time.Now().Format("2006-01-02 15:04:05.000")

	// c.IP()

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

	/*
	   db, err := connectToSQLite()
	   if err != nil {
	       log.Fatal(err)
	   }
	   defer db.Close()
	*/

	pid := os.Getpid()
	//pid := strconv.Itoa(os.Getpid())

	dt := time.Now().Format("2006-01-02 15:04:05.000")

	// c.IP()

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
