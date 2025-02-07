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
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Fprint(os.Stdout, "$ ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" {
			break
		}

		parts := strings.Fields(input)
		cmdName := parts[0]
		args := parts[1:]

		_, err := exec.LookPath(cmdName)
		if err != nil {
			fmt.Fprintf(os.Stderr, cmdName, "command not found: %s\n")
			continue
		}

		cmd := exec.Command(cmdName, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}

	}
}
