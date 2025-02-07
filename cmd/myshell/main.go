package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var _ = fmt.Fprint

func main() {
	fmt.Fprint(os.Stdout, "$ ")

	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input: ", err)
		os.Exit(1)
	}

	command := strings.TrimSpace(input)
	parts := strings.Fields(command)
	if len(parts) == 0 {
		os.Exit(0)
	}

	cmdName := parts[0]
	_, err = exec.LookPath(cmdName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
		os.Exit(1)
	}

	os.Exit(0)
}
