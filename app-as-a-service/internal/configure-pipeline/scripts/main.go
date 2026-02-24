package main

import (
	"fmt"
	"log"
	"os"

	kratix "github.com/syntasso/kratix-go"
	"github.com/syntasso/kratix-marketplace/app/tasks"
)

func main() {
	sdk := kratix.New()
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalf("usage: %s <pipeline-name>", os.Args[0])
	}

	pipelineName := args[0]
	fmt.Println("Pipeline requested:", pipelineName)

	st, err := sdk.ReadStatus()
	if err != nil {
		st = kratix.NewStatus()
	}

	switch pipelineName {
	case "resource-configure":
		if err := tasks.ConfigureResource(sdk, st); err != nil {
			log.Fatalf("resource pipeline: %v", err)
		}
	case "database-configure":
		if err := tasks.ConfigureDatabase(sdk, st); err != nil {
			log.Fatalf("database pipeline: %v", err)
		}
	case "vault-configure":
		if err := tasks.ConfigureVault(sdk, st); err != nil {
			log.Fatalf("vault pipeline: %v", err)
		}
	case "wait-db-ready":
		if err := tasks.WaitDatabaseReady(sdk, st); err != nil {
			log.Fatalf("wait database ready pipeline: %v", err)
		}
	default:
		log.Fatalf("unknown pipeline %q", sdk.PipelineName())
	}

	fmt.Println("Finished executing main.")
}
