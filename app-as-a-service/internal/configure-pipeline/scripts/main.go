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

	pipeline := args[0]
	fmt.Println("Pipeline requested:", pipeline)

	st, err := sdk.ReadStatus()
	if err != nil {
		st = kratix.NewStatus()
	}

	switch pipeline {
	case "promise-configure":
		err = tasks.PromiseConfigure(sdk)
	case "resource-configure":
		err = tasks.ResourceConfigure(sdk, st)
	case "database-configure":
		err = tasks.DatabaseConfigure(sdk, st)
	case "vault-configure":
		err = tasks.VaultConfigure(sdk, st)
	case "wait-db-ready":
		err = tasks.WaitDbReady(sdk, st)
	case "resource-delete":
		err = tasks.ResourceDelete(sdk)
	default:
		log.Fatalf("unknown pipeline %q", sdk.PipelineName())
	}

	if err != nil {
		log.Fatalf("%s pipeline: %v", pipeline, err)
	}

	fmt.Println("Finished executing main.")
}
