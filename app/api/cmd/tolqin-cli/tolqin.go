package main

import (
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/ztimes2/tolqin/app/api/internal/cli/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ oops... %s\n", err.Error())
		os.Exit(1)
	}
}
