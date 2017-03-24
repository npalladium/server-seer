package src

import (
	"../src/logger"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func GetFileContents(fileName string) []byte {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		ExitApplicationWithMessage(fmt.Sprintf("Failed to open a file: %s", err))
	}

	return data
}

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func ExitApplicationWithMessage(msg string) {
	if logger.Logger != nil {
		logger.Logger.Log(msg)
	} else {
		fmt.Fprintf(os.Stderr, "error: %v\n", msg)
	}
	os.Exit(1)
}

func RunCommand(command string) string {
	var (
		output []byte
		err    error
	)
	if output, err = exec.Command("/bin/sh", "-c", command).Output(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return strings.TrimSpace(string(output))
}
