package commands


import (
	"fmt"
	"os"
)

func exit(cmdline string) error {
	fmt.Printf("Goodbye!\n")
	os.Exit(0)
	return errRuntimeError
}