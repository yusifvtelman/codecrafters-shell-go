package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

// CommandHandler defines the signature for built-in command functions.
// The second parameter is the output filename for redirection (if any).
type CommandHandler func(args []string, outputFile string) error

// shellBuiltins is a separate map for the names of shell built-in commands.
var shellBuiltins = map[string]bool{
	"exit": true,
	"cd":   true,
	"echo": true,
	"pwd":  true,
	"type": true,
}

// commands maps command names to their handling functions.
var commands = map[string]CommandHandler{
	"exit": exitCommand,
	"cd":   cdCommand,
	"echo": echoCommand,
	"pwd":  pwdCommand,
	"type": typeCommand,
}

func exitCommand(args []string, outputFile string) error {
	// outputFile is ignored for exit.
	exitCode := 0
	if len(args) > 0 {
		_, err := fmt.Sscanf(args[0], "%d", &exitCode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "exit: %s: numeric argument required\n", args[0])
			exitCode = 1
		}
	}
	os.Exit(exitCode)
	return nil
}

func cdCommand(args []string, outputFile string) error {
	// outputFile is ignored for cd.
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "cd: missing argument")
		return nil
	}
	dir := args[0]
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
	return nil
}

func echoCommand(args []string, outputFile string) error {
	output := strings.Join(args, " ")
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			return err
		}
		defer file.Close()
		fmt.Fprintln(file, output)
	} else {
		fmt.Println(output)
	}
	return nil
}

func pwdCommand(args []string, outputFile string) error {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		return err
	}
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			return err
		}
		defer file.Close()
		fmt.Fprintln(file, dir)
	} else {
		fmt.Println(dir)
	}
	return nil
}

func typeCommand(args []string, outputFile string) error {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "type: missing argument")
		return nil
	}
	target := args[0]
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			return err
		}
		defer file.Close()
		// Use shellBuiltins instead of commands to avoid the init cycle.
		if shellBuiltins[target] {
			fmt.Fprintf(file, "%s is a shell builtin\n", target)
		} else if path, err := exec.LookPath(target); err == nil {
			fmt.Fprintf(file, "%s is %s\n", target, path)
		} else {
			fmt.Fprintf(file, "%s not found\n", target)
		}
	} else {
		if shellBuiltins[target] {
			fmt.Printf("%s is a shell builtin\n", target)
		} else if path, err := exec.LookPath(target); err == nil {
			fmt.Printf("%s is %s\n", target, path)
		} else {
			fmt.Printf("%s not found\n", target)
		}
	}
	return nil
}

func main() {
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

		// Process tokens and handle output redirection.
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
				i += 2 // Skip the operator and filename.
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

		// Execute built-in commands.
		if handler, ok := commands[cmdName]; ok {
			if err := handler(cmdArgs, outputFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error executing %s: %v\n", cmdName, err)
			}
			continue
		}

		// Execute external commands.
		path, err := exec.LookPath(cmdName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: command not found\n", cmdName)
			continue
		}
		cmd := exec.Command(path, cmdArgs...)
		cmd.Stdin = os.Stdin
		cmd.Stderr = os.Stderr
		if outputFile != "" {
			file, err := os.Create(outputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
				continue
			}
			cmd.Stdout = file
			if err := cmd.Run(); err != nil {
				file.Close()
				if exitErr, ok := err.(*exec.ExitError); ok {
					os.Exit(exitErr.ExitCode())
				}
				os.Exit(1)
			}
			file.Close()
		} else {
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				continue
			}
		}
	}
}

// tokenize breaks the input into tokens, handling quotes and escape sequences.
func tokenize(input string) []string {
	var tokens []string
	var currentToken []rune
	inDoubleQuote := false
	inSingleQuote := false
	escapeNext := false

	for i := 0; i < len(input); i++ {
		c := rune(input[i])

		if escapeNext {
			currentToken = append(currentToken, c)
			escapeNext = false
			continue
		}

		if inDoubleQuote {
			if c == '\\' {
				escapeNext = true
			} else if c == '"' {
				inDoubleQuote = false
			} else {
				currentToken = append(currentToken, c)
			}
			continue
		}

		if inSingleQuote {
			if c == '\'' {
				inSingleQuote = false
			} else {
				currentToken = append(currentToken, c)
			}
			continue
		}

		// Check for "1>" operator.
		if c == '1' && i+1 < len(input) && input[i+1] == '>' {
			if len(currentToken) > 0 {
				tokens = append(tokens, string(currentToken))
				currentToken = nil
			}
			tokens = append(tokens, "1>")
			i++ // Skip the '>' character.
			continue
		}

		if c == '>' {
			if len(currentToken) > 0 {
				tokens = append(tokens, string(currentToken))
				currentToken = nil
			}
			tokens = append(tokens, ">")
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' {
			inDoubleQuote = true
			continue
		}

		if c == '\'' {
			inSingleQuote = true
			continue
		}

		if unicode.IsSpace(c) {
			if len(currentToken) > 0 {
				tokens = append(tokens, string(currentToken))
				currentToken = nil
			}
			continue
		}

		currentToken = append(currentToken, c)
	}

	if len(currentToken) > 0 {
		tokens = append(tokens, string(currentToken))
	}

	return tokens
}
