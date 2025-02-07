package main

import (
	"bufio"
	"fmt"
	"os"
)

var _ = fmt.Fprint

func main() {
	fmt.Fprint(os.Stdout, "$ ")
	bufio.NewReader(os.Stdin).ReadString('\n')
}
