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

		// Process redirection and extract command arguments
		var args []string
		var outputFile string
		syntaxError := false

		for i := 0; i < len(tokens); {
			token := tokens[i]
			if token == ">" || token == "1>" {
				if i+1 >= len(tokens) {
					fmt.Fprintln(os.Stderr, "Syntax error: no filename provided for redirection.")
					syntaxError = true
					break
				}
				outputFile = tokens[i+1]
				i += 2 // Skip operator and filename
			} else {
				args = append(args, token)
				i++
			}
		}

		if syntaxError || len(args) == 0 {
			continue
		}

		cmdName := args[0]
		cmdArgs := args[1:]

		// Handle built-in commands
		if cmdName == "exit" {
			exitCode := 0
			if len(cmdArgs) > 0 {
				_, err := fmt.Sscanf(cmdArgs[0], "%d", &exitCode)
				if err != nil {
					fmt.Fprintf(os.Stderr, "exit: %s: numeric argument required\n", cmdArgs[0])
					exitCode = 1
				}
			}
			os.Exit(exitCode)
		}

		if cmdName == "echo" {
			fmt.Println(strings.Join(cmdArgs, " "))
			continue
		}

		if cmdName == "type" {
			if len(cmdArgs) == 0 {
				fmt.Fprintln(os.Stderr, "type: missing argument")
				continue
			}

			target := cmdArgs[0]
			if builtinMap[target] {
				fmt.Printf("%s is a shell builtin\n", target)
			} else if path, err := exec.LookPath(target); err == nil {
				fmt.Printf("%s is %s\n", target, path)
			} else {
				fmt.Printf("%s not found\n", target)
			}
			continue
		}

		if cmdName == "pwd" {
			if dir, err := os.Getwd(); err == nil {
				fmt.Println(dir)
			}
			continue
		}

		if cmdName == "cd" {
			if len(cmdArgs) == 0 {
				fmt.Fprintln(os.Stderr, "cd: missing argument")
				continue
			}

			dir := cmdArgs[0]
			if strings.HasPrefix(dir, "~/") {
				if home, err := os.UserHomeDir(); err == nil {
					dir = home + dir[1:]
				}
			} else if dir == "~" {
				if home, err := os.UserHomeDir(); err == nil {
					dir = home
				}
			}

			if err := os.Chdir(dir); err != nil {
				fmt.Fprintf(os.Stderr, "cd: %s: %v\n", dir, err)
			}
			continue
		}

		// Handle external commands with redirection
		path, err := exec.LookPath(cmdName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
			continue
		}

		cmd := exec.Command(path, cmdArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr

		// Handle output redirection
		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				continue
			}
			defer file.Close()
			cmd.Stdout = file
		} else {
			cmd.Stdout = os.Stdout
		}

		if err := cmd.Run(); err != nil {
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
	inDoubleQuote := false
	inSingleQuote := false
	escapeNext := false

	for i, r := range input {
		if escapeNext {
			currentToken = append(currentToken, r)
			escapeNext = false
			continue
		}

		if inDoubleQuote {
			switch r {
			case '\\':
				escapeNext = true
			case '"':
				inDoubleQuote = false
			default:
				currentToken = append(currentToken, r)
			}
			continue
		}

		if inSingleQuote {
			if r == '\'' {
				inSingleQuote = false
			} else {
				currentToken = append(currentToken, r)
			}
			continue
		}

		switch r {
		case '\\':
			escapeNext = true
		case '\'':
			inSingleQuote = true
		case '"':
			inDoubleQuote = true
		case '>':
			if len(currentToken) > 0 {
				tokens = append(tokens, string(currentToken))
				currentToken = nil
			}
			tokens = append(tokens, ">")
		case '1':
			if i+1 < len(input) && input[i+1] == '>' {
				if len(currentToken) > 0 {
					tokens = append(tokens, string(currentToken))
					currentToken = nil
				}
				tokens = append(tokens, "1>")
				i++ // Skip next character
				continue
			}
			currentToken = append(currentToken, r)
		default:
			if unicode.IsSpace(r) {
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
