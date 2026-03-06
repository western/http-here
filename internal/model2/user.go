package model2

import (
	"encoding/json"
	"fmt"
	//"os"
	//"regexp"
	"strconv"
	//"strings"
	"time"

	//"github.com/western/http-here/v2/internal/conf"
	"github.com/western/http-here/v2/internal/util"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/jedib0t/go-pretty/v6/table"
	//"github.com/jedib0t/go-pretty/v6/text"
)

type User struct {
	ID uint64 `json:"id"`

	Login    string `json:"login"`
	Password string `json:"password"`

	Enabled bool   `json:"status"`
	Label   string `json:"label"`

	Registered string `json:"registered"`
	Changed    string `json:"changed"`
}

func UserFindByLogin(arg_login string) (User, bool) {

	if len(arg_login) == 0 {
		panic("Yous should set login")
	}

	var user User

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {

			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("user_")

			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				item.Value(func(v []byte) error {

					var el User
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						fmt.Println("error:", err2)
					}

					if el.Login == arg_login {

						//fmt.Printf("key=%s, value=%+v\n", k, el)
						user = el
					}

					return nil
				})
			}
			return nil
		})

		if user.ID == 0 {
			return user, false
		} else {
			return user, true
		}

	}

	return user, false
}

func UserSave(user User) error {

	if user.ID == 0 {
		panic("Yous should set User")
	}

	var key string

	if Dbse.BadgerEnable {

		_, isFound := BadgerGetOne("user_" + strconv.FormatUint(user.ID, 10))
		if isFound {
			key = "user_" + strconv.FormatUint(user.ID, 10)
		}

		if len(key) > 0 {

			_dt := time.Now().Format("2006-01-02 15:04:05.000")
			user.Changed = _dt

			payload, _ := json.Marshal(user)

			Dbse.Badger.Update(func(txn *badger.Txn) error {

				err := txn.Set([]byte(key), []byte(payload))
				if err != nil {
					panic(err)
				}

				return nil
			})
		}

	}

	return nil
}

func UserGetPrimaryId() uint64 {

	if !Dbse.BadgerEnable {
		return 0
	}

	seq, err := Dbse.Badger.GetSequence([]byte("seq_user"), 1000)
	if err != nil {
		panic(err)
	}
	defer seq.Release()

	// uint64, err
	num, err := seq.Next()
	if err != nil {
		panic(err)
	}

	if num == 0 {
		num, err = seq.Next()
		if err != nil {
			panic(err)
		}
	}

	return num
}

func UserAdd(arg_login, arg_password, arg_label string, arg_disabled bool) {

	if Dbse.BadgerEnable {

		user_id := UserGetPrimaryId()
		login := "login" + strconv.FormatUint(user_id, 10) + util.RandStringRunes(2)
		password := util.RandStringRunes(16)

		if len(arg_login) > 0 {

			_, isFound := BadgerGetOne("idx_user_login_" + arg_login)
			if isFound {
				panic("Login should be unique in system")
			}

			login = arg_login
		}
		if len(arg_password) > 0 {
			password = arg_password
		}

		_dt := time.Now().Format("2006-01-02 15:04:05.000")
		//_dt_timestamp, _ := time.Parse("2006-01-02 15:04:05.000", _dt)

		el := User{
			ID: user_id,

			Login:    login,
			Password: password,

			Enabled: !arg_disabled,
			Label:   arg_label,

			Registered: _dt,
			Changed:    "",
		}
		payload, _ := json.Marshal(el)

		key := "user_"

		//_dt_timestamp_str := strconv.FormatInt(_dt_timestamp.Unix(), 10)
		//key += _dt_timestamp_str
		key += strconv.FormatUint(user_id, 10)

		Dbse.Badger.Update(func(txn *badger.Txn) error {

			err := txn.Set([]byte(key), []byte(payload))
			if err != nil {
				panic(err)
			}

			return nil
		})

		UserList()

	}
}

func UserGenerate() {

	if Dbse.BadgerEnable {

		for i := range 10 {

			//i := i + 1
			_ = i

			user_id := UserGetPrimaryId()
			login := "login" + strconv.FormatUint(user_id, 10) + util.RandStringRunes(2)
			password := util.RandStringRunes(16)

			_dt := time.Now().Format("2006-01-02 15:04:05.000")
			//_dt_timestamp, _ := time.Parse("2006-01-02 15:04:05.000", _dt)

			el := User{
				ID: user_id,

				Login:    login,
				Password: password,

				Enabled: true,
				Label:   "",

				Registered: _dt,
				Changed:    "",
			}
			payload, _ := json.Marshal(el)

			key := "user_"

			//_dt_timestamp_str := strconv.FormatInt(_dt_timestamp.Unix()+int64(i), 10)
			//key += _dt_timestamp_str
			key += strconv.FormatUint(user_id, 10)

			Dbse.Badger.Update(func(txn *badger.Txn) error {

				err := txn.Set([]byte(key), []byte(payload))
				if err != nil {
					panic(err)
				}

				err = txn.Set([]byte("idx_user_login_"+login), []byte(key))
				if err != nil {
					panic(err)
				}

				return nil
			})

		}

		UserList()

	}

}

func UserList() {

	t := table.NewWriter()
	//t.SetOutputMirror(os.Stdout)
	//t.SetStyle(table.StyleBold)
	t.SetStyle(table.StyleLight)

	// key=user_2, value={ID:2 Login:login2DP Password:SgJQY6pk2iNWsUxQ Enabled:true Label: Registered:2026-03-05 23:19:59.559 Changed:}
	t.AppendHeader(table.Row{"#", "Login", "Password", "Enabled", "Label", "Registered", "Changed"})

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("user_")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				err := item.Value(func(v []byte) error {

					//fmt.Printf("key=%s, value=%s\n", k, v)

					var el User
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						fmt.Println("error:", err2)
					}

					//fmt.Printf("key=%s, value=%+v\n", k, el)
					t.AppendRows([]table.Row{
						{el.ID, el.Login, el.Password, el.Enabled, el.Label, el.Registered, el.Changed},
					})

					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

		fmt.Println(t.Render())

	}
}

func UserListFill() map[string]string {

	var password_list = map[string]string{}

	if Dbse.BadgerEnable {

		Dbse.Badger.View(func(txn *badger.Txn) error {
			it := txn.NewIterator(badger.DefaultIteratorOptions)
			defer it.Close()
			prefix := []byte("user_")
			for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
				item := it.Item()
				//k := item.Key()
				err := item.Value(func(v []byte) error {

					//fmt.Printf("key=%s, value=%s\n", k, v)

					var el User
					err2 := json.Unmarshal(v, &el)
					if err2 != nil {
						fmt.Println("error:", err2)
					}

					//fmt.Printf("key=%s, value=%+v\n", k, el)

					if el.Enabled {
						password_list[el.Login] = el.Password
					}

					return nil
				})
				if err != nil {
					return err
				}
			}
			return nil
		})

	}

	return password_list
}

func UserClear() {

	if Dbse.BadgerEnable {

		// -------------------------------------------------------------------------------------------------------------------------------------------

		prefix := []byte("user_")

		err := Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

		// -------------------------------------------------------------------------------------------------------------------------------------------

		prefix = []byte("idx_user_login_")

		err = Dbse.Badger.DropPrefix(prefix)
		if err != nil {
			panic(err)
		}

	}
}
