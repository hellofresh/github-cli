package main

import (
	"fmt"
	"os"

	"github.com/hellofresh/github-cli/cmd"
)

func main() {
	if err := cmd.NewRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
