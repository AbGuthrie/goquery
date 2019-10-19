package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

func clear(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) > 1 {
		return fmt.Errorf("This command takes no parameters")
	}

	// TODO: it may make more sense for the os runtime to be set and retrieved
	// via the config/state.go module
	currentOS := runtime.GOOS
	if currentOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		print("\033[H\033[2J")
	}
	return nil
}

func clearHelp() string {
	return "Clear the terminal screen"
}

func clearSuggest(cmdline string) []prompt.Suggest {
	return []prompt.Suggest{}
}
