package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	builtin := []string{"exit", "echo", "type", "pwd", "cd"}
	builtinMap := make(map[string]bool)
	for _, b := range builtin {
		builtinMap[b] = true
	}

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

		// If i Swich case ile değiştir.

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

		if cmdName == "type" {
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "type: missing argument")
				continue
			}

			if builtinMap[args[0]] {
				fmt.Println(args[0], "is a shell builtin")
				continue
			}

			path, err := exec.LookPath(args[0])
			if err == nil {
				fmt.Printf("%s is %s\n", args[0], path)
				continue
			}

			fmt.Printf("%s not found\n", args[0])
			continue
		}

		if cmdName == "pwd" {
			dir, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(dir)
			continue
		}

		if cmdName == "cd" {
			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "cd: missing argument")
				continue
			}

			if args[0] == "~" {
				args[0] = os.Getenv("HOME")
			}

			err := os.Chdir(args[0])
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Fprintf(os.Stderr, "cd: %s: No such file or directory\n", args[0])
				} else {
					fmt.Fprintf(os.Stderr, "cd: %s: %v\n", args[0], err)
				}
			}
			continue
		}

		path, err := exec.LookPath(cmdName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
			continue
		}

		cmd := exec.Command(path, args...)
		cmd.Args[0] = cmdName
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()

		if err != nil {
			fmt.Printf("%s: command not found\n", cmdName)
		}
	}
}

func exit() {
	os.Exit(0)
}
