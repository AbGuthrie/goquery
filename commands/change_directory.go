package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AbGuthrie/goquery/v2/config"
	"github.com/AbGuthrie/goquery/v2/hosts"
	"github.com/AbGuthrie/goquery/v2/models"
	"github.com/AbGuthrie/goquery/v2/utils"

	prompt "github.com/c-bata/go-prompt"
)

var verificationTemplate = "select * from file where path = '%s' and type = 'directory'"

func changeDirectory(api models.GoQueryAPI, config *config.Config, cmdline string) error {
	host, err := hosts.GetCurrentHost()
	if err != nil {
		return fmt.Errorf("No host is currently connected: %s", err)
	}

	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) == 1 {
		return fmt.Errorf("Directory must be provided")
	}

	// TODO This needs to support Unicode/Runes
	// TODO Support fast mode that doesn't do directory verification
	requestedDirectory := cmdline[strings.Index(cmdline, " ")+1:]

	if len(requestedDirectory) == 0 {
		return fmt.Errorf("Directory requested is invalid")
	}

	// The change isn't absolute so we need the current directory
	if requestedDirectory[0] != '/' {
		requestedDirectory = host.CurrentDirectory + requestedDirectory
	}
	requestedDirectory = filepath.Clean(requestedDirectory)

	// All directory changes must end with a forward slash
	if requestedDirectory[len(requestedDirectory)-1] != '/' {
		requestedDirectory += "/"
	}

	verificationQuery := fmt.Sprintf(verificationTemplate, requestedDirectory)
	results, err := utils.ScheduleQueryAndWait(api, host.UUID, verificationQuery)

	if err != nil {
		return err
	}

	if len(results) != 1 {
		return fmt.Errorf("No such directory")
	}

	return hosts.SetCurrentHostDirectory(requestedDirectory)
}

func changeDirectoryHelp() string {
	return "Change directories on a remote host"
}

func changeDirectorySuggest(cmdline string) []prompt.Suggest {
	return []prompt.Suggest{}
}
