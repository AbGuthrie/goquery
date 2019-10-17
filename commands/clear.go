package commands

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func clear(cmdline string) error {
	args := strings.Split(cmdline, " ") // Separate command and arguments
	if len(args) > 1 {
		return fmt.Errorf("This commands takes no parameters")
	}

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
