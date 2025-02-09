package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

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

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		cmdName := parts[0]
		args := parts[1:]

		if cmdName == "exit" {
			exitCode := 0
			if len(args) > 0 {
				_, err := fmt.Sscanf(args[0], "%d", &exitCode)
				if err != nil {
					exitCode = 0
				}
			}
			os.Exit(exitCode)
		}

		if cmdName == "echo" {
			fmt.Println(strings.Join(args, " "))
			continue
		}

		path, err := exec.LookPath(cmdName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
			continue
		}

		cmd := exec.Command(path, args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}
}
