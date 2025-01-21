package main

import (
	"fmt"
	"os"

	commands "github.com/base-go/cmd/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
