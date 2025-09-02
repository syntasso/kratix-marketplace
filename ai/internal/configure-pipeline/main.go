package main

import (
	"log"

	kratix "github.com/syntasso/kratix-go"
)

var sdk = kratix.New()

func main() {
	log.Printf("Workflow action: %s", sdk.WorkflowAction())
	log.Printf("Workflow type: %s", sdk.WorkflowType())
	log.Printf("Promise name: %s", sdk.PromiseName())
	log.Printf("Pipeline name: %s", sdk.PipelineName())

	if sdk.WorkflowType() == "promise" {
		if sdk.PipelineName() == "provision-postgres-db" {
			deployPostgres()
			return
		}
		if sdk.PipelineName() != "litellm-deploy" {
			deployLiteLLM()
			return
		}
		panic("unknown pipeline name " + sdk.PipelineName())
	} else if sdk.WorkflowType() == "resource" {
		res, err := sdk.ReadResourceInput()
		if err != nil {
			panic(err)
		}
		team, err := res.GetValue("spec.team")
		if err != nil {
			panic(err)
		}
		tier, err := res.GetValue("spec.tier")
		if err != nil {
			panic(err)
		}
		models, err := res.GetValue("spec.models")
		if err != nil {
			panic(err)
		}
		log.Printf("Setting up LiteLLM for team %s on tier %s with models %v", team, tier, models)
		setupLiteLLMTeam(tier.(string), team.(string), toStringSlice(models))

		uiEnabled, err := res.GetValue("spec.ui")
		if err != nil {
			panic(err)
		}
		if uiEnabledBool, ok := uiEnabled.(bool); ok && uiEnabledBool {
			log.Printf("Setting up OpenWebUI")
			deployOpenWebUI()
		}
	} else {
		panic("unknown workflow type " + sdk.WorkflowType())
	}
}

func toStringSlice(input any) []string {
	if input == nil {
		return []string{}
	}
	interfaceSlice, ok := input.([]any)
	if !ok {
		panic("expected []any but got different type")
	}
	stringSlice := make([]string, len(interfaceSlice))
	for i, v := range interfaceSlice {
		str, ok := v.(string)
		if !ok {
			panic("expected string but got different type")
		}
		stringSlice[i] = str
	}
	return stringSlice
}
