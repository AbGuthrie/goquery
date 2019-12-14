package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AbGuthrie/goquery/config"
	"github.com/AbGuthrie/goquery/hosts"
	"github.com/AbGuthrie/goquery/models"
	"github.com/AbGuthrie/goquery/utils"

	prompt "github.com/c-bata/go-prompt"
)

func listDirectory(api models.GoQueryAPI, config config.Config, cmdline string) error {
	host, err := hosts.GetCurrentHost()
	if err != nil {
		return fmt.Errorf("No host is currently connected: %s", err)
	}

	args := strings.Split(cmdline, " ") // Separate command and arguments
	lsDir := "."
	if len(args) >= 2 {
		lsDir = cmdline[len(args[0])+1:]
		if len(lsDir) == 0 {
			return fmt.Errorf("Invalid Directory")
		}
	}

	if lsDir[0] != '/' {
		lsDir = host.CurrentDirectory + lsDir
	}
	lsDir = filepath.Clean(lsDir)

	// All directory changes must end with a forward slash
	if lsDir[len(lsDir)-1] != '/' {
		lsDir += "/"
	}

	listQuery := fmt.Sprintf("select * from file where directory = '%s'", lsDir)
	results, err := utils.ScheduleQueryAndWait(api, host.UUID, listQuery)

	if err != nil {
		return err
	}

	utils.PrettyPrintQueryResults(results, config.PrintMode)
	return nil
}

func listDirectoryHelp() string {
	return "List the files in the current directory on the remote host"
}

func listDirectorySuggest(cmdline string) []prompt.Suggest {
	return []prompt.Suggest{}
}
