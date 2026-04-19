package app

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/western/http-here/v2/internal/model2"
)

func cmdSubcommandUser() {

	userCmd := flag.NewFlagSet("user", flag.ExitOnError)
	arg_help := userCmd.Bool("help", false, "Show help")

	arg_login := userCmd.String("login", "", "Login for user basic auth")
	arg_password := userCmd.String("password", "", "Password for user basic auth")
	arg_label := userCmd.String("label", "", "Label for user account")
	arg_disabled := userCmd.Bool("disabled", false, "Disabled for user account")

	arg_generate := userCmd.Bool("generate", false, "Generate list of random accounts")
	arg_disable_all := userCmd.Bool("disable-all", false, "Disable all User accounts")
	arg_enable_all := userCmd.Bool("enable-all", false, "Enable all User accounts")
	arg_list := userCmd.Bool("list", false, "Print all user accounts")
	arg_truncate := userCmd.Bool("truncate", false, "Recreate User storage with drop all users")

	userCmd.Parse(os.Args[2:])

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_help {

		inf := []string{
			``,
			``,
			`usage: ` + green_clr(`http-here user`) + ` [options] `,
			``,
			`options:`,
			``,
			`     --login ` + white_clr(`[str]`) + `               Login for NEW USER basic authorization`,
			`     --password ` + white_clr(`[str]`) + `            Password for NEW USER basic authorization`,
			`     --label ` + white_clr(`[str]`) + `               Label for NEW USER account`,
			`     --disabled                  Disabled for NEW USER account`,
			``,
			`     --generate                  Generate and save list of random accounts`,
			`     --disable-all               Disable all User accounts`,
			`     --enable-all                Enable all User accounts`,
			``,
			`     --list                      Print all user accounts`,
			``,
			`     --truncate                  Recreate User storage with drop all users`,
			``,
			``,
			`examples:`,
			``,
			`     Add user with login (password will be generate as random)`,
			`                        ` + green_clr(`http-here user`) + ` --login ` + white_clr(`vasya3000`),
			``,
			`     Add user with login and password`,
			`                        ` + green_clr(`http-here user`) + ` --login ` + white_clr(`vasya3001`) + ` --password ` + white_clr(`pass3001`),
			``,
			``,
			`also:`,
			`       ` + green_clr(`http-here usermod`) + ` --help `,
			`       ` + green_clr(`http-here userdel`) + ` --help `,
			``,
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if len(*arg_login) > 0 {

		model2.UserAdd(*arg_login, *arg_password, *arg_label, *arg_disabled)
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_generate {

		model2.UserGenerate()
		os.Exit(0)
	}

	if *arg_disable_all {

		model2.UserSetStatusAll(false)
		os.Exit(0)
	}

	if *arg_enable_all {

		model2.UserSetStatusAll(true)
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_list {

		model2.UserList()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_truncate {

		model2.UserTruncate()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}

func cmdSubcommandUserMod() {

	usermodCmd := flag.NewFlagSet("usermod", flag.ExitOnError)
	arg_help := usermodCmd.Bool("help", false, "Show help")

	arg_login := usermodCmd.String("login", "", "Login")
	//arg_newlogin := usermodCmd.String("newlogin", "", "New login for account")
	arg_password := usermodCmd.String("password", "", "Password for account")
	arg_label := usermodCmd.String("label", "", "Label for account")
	arg_disable := usermodCmd.Bool("disable", false, "Disable for account")
	arg_enable := usermodCmd.Bool("enable", false, "Enable for account")

	usermodCmd.Parse(os.Args[2:])

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_help {

		inf := []string{
			``,
			``,
			`usage: ` + green_clr(`http-here usermod`) + ` [options] `,
			``,
			`options:`,
			``,
			`     --login ` + white_clr(`[str]`) + `                Login`,
			//`     --newlogin ` + white_clr(`[str]`) + `             New login for account`,
			`     --password ` + white_clr(`[str]`) + `             Password for account`,
			`     --label ` + white_clr(`[str]`) + `                Label for account`,
			``,
			`     --disable                    Disable for account`,
			`     --enable                     Enable for account`,
			``,
			``,
			`examples:`,
			``,
			`     Disable user`,
			`                        ` + green_clr(`http-here usermod`) + ` --login ` + white_clr(`XXXXXXX`) + ` --disable`,
			``,
			`     Change password`,
			`                        ` + green_clr(`http-here usermod`) + ` --login ` + white_clr(`XXXXXXX`) + ` --password ` + white_clr(`TTTTTTTTT`),
			``,
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if len(*arg_login) == 0 {

		fmt.Println("")
		fmt.Println("Login is mandatory param")
		os.Exit(0)
	}

	user, isFound := model2.UserFindByLogin(*arg_login)

	if !isFound {

		//fmt.Println(`User "` + *arg_login + `" nof found`)
		os.Exit(0)

	} else {

		/*
			if len(*arg_newlogin) > 0 {
				user.Login = *arg_newlogin
			}
		*/

		if len(*arg_password) > 0 {
			user.Password = *arg_password
		}

		if len(*arg_label) > 0 {
			user.Label = *arg_label
		}

		if *arg_disable {
			user.Enabled = false
		}

		if *arg_enable {
			user.Enabled = true
		}

		model2.UserUpdate(user)
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}

func cmdSubcommandUserDel() {

	usermodCmd := flag.NewFlagSet("userdel", flag.ExitOnError)
	arg_help := usermodCmd.Bool("help", false, "Show help")

	arg_login := usermodCmd.String("login", "", "Login")

	usermodCmd.Parse(os.Args[2:])

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_help {

		inf := []string{
			``,
			``,
			`usage: ` + green_clr(`http-here userdel`) + ` [options] `,
			``,
			`options:`,
			``,
			`     --login ` + white_clr(`[str]`) + `                Login for delete`,
			``,
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if len(*arg_login) == 0 {

		fmt.Println("")
		fmt.Println("Login is mandatory param")
		os.Exit(0)
	}

	user, isFound := model2.UserFindByLogin(*arg_login)

	if !isFound {

		fmt.Println(`User "` + *arg_login + `" nof found`)
		os.Exit(0)

	} else {

		model2.UserDelByLogin(user.Login)
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}

func cmdSubcommandLog() {

	logCmd := flag.NewFlagSet("log", flag.ExitOnError)
	arg_help := logCmd.Bool("help", false, "Show help")

	arg_dumpto := logCmd.String("dumpto", "", "Filename for dump")
	arg_jsonto := logCmd.String("jsonto", "", "Filename for json dump")
	arg_dump := logCmd.Bool("dump", false, "Dump to stdout")

	arg_truncate := logCmd.Bool("truncate", false, "Recreate Log storage with drop all records")

	logCmd.Parse(os.Args[2:])

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_help {

		inf := []string{
			``,
			``,
			`usage: ` + green_clr(`http-here log`) + ` [options] `,
			``,
			`options:`,
			``,
			`     --dumpto ` + white_clr(`[str]`) + `              Filename for dump`,
			`     --jsonto ` + white_clr(`[str]`) + `              Filename for json dump`,
			`     --dump                      Dump to stdout`,
			``,
			`     --truncate                  Recreate Log storage with drop all records`,

			``,
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if len(*arg_dumpto) > 0 {

		model2.EventLogDumpTo(*arg_dumpto)
		os.Exit(0)
	}

	if len(*arg_jsonto) > 0 {

		model2.EventLogJsonTo(*arg_jsonto)
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_dump {

		model2.EventLogDump()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_truncate {

		model2.EventLogTruncate()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}

func cmdSubcommandFile() {

	fileCmd := flag.NewFlagSet("file", flag.ExitOnError)
	arg_help := fileCmd.Bool("help", false, "Show help")

	arg_list := fileCmd.Bool("list", false, "Print all database files")
	arg_truncate := fileCmd.Bool("truncate", false, "Recreate File storage with drop all records")

	fileCmd.Parse(os.Args[2:])

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_help {

		inf := []string{
			``,
			``,
			`usage: ` + green_clr(`http-here file`) + ` [options] `,
			``,
			`options:`,
			``,
			`     --list                      Print all database files`,
			``,
			`     --truncate                  Recreate File storage with drop all records`,

			``,
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_list {

		model2.FileList()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_truncate {

		model2.FileTruncate()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}

func cmdSubcommandDatabase() {

	databaseCmd := flag.NewFlagSet("database", flag.ExitOnError)
	arg_help := databaseCmd.Bool("help", false, "Show help")

	arg_dumpto := databaseCmd.String("dumpto", "", "Filename for dump")

	arg_restore := databaseCmd.String("restore", "", "Filename for restore")

	arg_destroy := databaseCmd.Bool("destroy", false, "Destroy database")

	databaseCmd.Parse(os.Args[2:])

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_help {

		inf := []string{
			``,
			``,
			`usage: ` + green_clr(`http-here database`) + ` [options] `,
			``,
			`options:`,
			``,
			`     --dumpto ` + white_clr(`[str]`) + `              Filename for dump`,
			``,
			``,
			`     --restore ` + white_clr(`[str]`) + `              Filename for restore`,
			``,
			``,
			`     --destroy                  Destroy database`,
			``,
			``,
		}

		fmt.Println(strings.Join(inf[:], "\n"))
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if len(*arg_dumpto) > 0 {

		model2.BadgerDumpTo(*arg_dumpto)
		os.Exit(0)
	}

	/*
		if len(*arg_jsonto) > 0 {

			model2.BadgerJsonTo(*arg_jsonto)
			os.Exit(0)
		}
	*/

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if len(*arg_restore) > 0 {

		model2.BadgerRestoreFrom(*arg_restore)
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_destroy {

		model2.BadgerDestroy()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}
