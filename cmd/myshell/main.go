package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
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
		line := scanner.Text()
		tokens := tokenize(line)

		if len(tokens) == 0 {
			continue
		}

		cmdName := tokens[0]
		args := tokens[1:]

		if cmdName == "exit" {
			exitCode := 0
			if len(args) > 0 {
				_, err := fmt.Sscanf(args[0], "%d", &exitCode)
				if err != nil {
					fmt.Fprintf(os.Stderr, "exit: %s: numeric argument required\n", args[0])
					exitCode = 1
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

			dir := args[0]
			if strings.HasPrefix(dir, "~/") {
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Fprintln(os.Stderr, "cd:", err)
					continue
				}
				dir = home + dir[1:]
			} else if dir == "~" {
				home, err := os.UserHomeDir()
				if err != nil {
					fmt.Fprintln(os.Stderr, "cd:", err)
					continue
				}
				dir = home
			}

			err := os.Chdir(dir)
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
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			os.Exit(1)
		}
	}
}

func tokenize(input string) []string {
	var tokens []string
	var currentToken []rune
	inSingleQuote := false

	for _, r := range input {
		if inSingleQuote {
			if r == '\'' {
				inSingleQuote = false
			} else {
				currentToken = append(currentToken, r)
			}
		} else {
			if r == '\'' {
				inSingleQuote = true
			} else if unicode.IsSpace(r) {
				if len(currentToken) > 0 {
					tokens = append(tokens, string(currentToken))
					currentToken = nil
				}
			} else {
				currentToken = append(currentToken, r)
			}
		}
	}

	if len(currentToken) > 0 {
		tokens = append(tokens, string(currentToken))
	}

	return tokens
}
