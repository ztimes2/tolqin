package main

import (
	"log"

	"github.com/ztimes2/tolqin/app/api/internal/cmd"
)

func main() {
	if err := cmd.New().Execute(); err != nil {
		log.Fatalf("❌ %s", err.Error())
	}
}
