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
	arg_list := userCmd.Bool("list", false, "Print all user accounts")
	arg_clear := userCmd.Bool("clear", false, "Clear all user accounts")

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
			`     --list                      Print all user accounts`,
			``,
			`     --clear                     Clear all user accounts`,

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

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_list {

		model2.UserList()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_clear {

		model2.UserClear()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}

func cmdSubcommandUserMod() {

	usermodCmd := flag.NewFlagSet("usermod", flag.ExitOnError)
	arg_help := usermodCmd.Bool("help", false, "Show help")

	arg_login := usermodCmd.String("login", "", "Login")
	arg_newlogin := usermodCmd.String("newlogin", "", "New login for account")
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
			`     --newlogin ` + white_clr(`[str]`) + `             New login for account`,
			`     --password ` + white_clr(`[str]`) + `             Password for account`,
			`     --label ` + white_clr(`[str]`) + `                Label for account`,
			``,
			`     --disable                    Disable for account`,
			`     --enable                     Enable for account`,
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
		return
	}

	user, isFound := model2.UserFindByLogin(*arg_login)

	if !isFound {

		fmt.Println(`User "` + *arg_login + `" nof found`)
		os.Exit(0)

	} else {

		if len(*arg_newlogin) > 0 {
			user.Login = *arg_newlogin
		}

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

		model2.UserSave(user)
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}

func cmdSubcommandLog() {

	logCmd := flag.NewFlagSet("log", flag.ExitOnError)
	arg_help := logCmd.Bool("help", false, "Show help")

	arg_dumpto := logCmd.String("dumpto", "", "Dump to file")
	arg_dump := logCmd.Bool("dump", false, "Dump to stdout")

	arg_clear := logCmd.Bool("clear", false, "Clear event log")

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
			`     --dump                      Dump to stdout`,
			``,
			`     --clear                     Clear event log`,

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

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_dump {

		model2.EventLogDump()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	if *arg_clear {

		model2.EventLogClear()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}

func cmdSubcommandFile() {

	fileCmd := flag.NewFlagSet("user", flag.ExitOnError)
	arg_help := fileCmd.Bool("help", false, "Show help")

	arg_list := fileCmd.Bool("list", false, "Print all user accounts")
	arg_clear := fileCmd.Bool("clear", false, "Clear all user accounts")

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
			`     --clear                     Clear all files`,

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

	if *arg_clear {

		model2.FileClear()
		os.Exit(0)
	}

	// -------------------------------------------------------------------------------------------------------------------------------------------

	os.Exit(0)

}
