package main

import (
	"fmt"
	"log"

	kratix "github.com/syntasso/kratix-go"
	kubernetes "k8s.io/client-go/kubernetes"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/client/config"
)

var sdk = kratix.New()

func main() {
	log.Printf("Workflow action: %s", sdk.WorkflowAction())
	log.Printf("Workflow type: %s", sdk.WorkflowType())
	log.Printf("Promise name: %s", sdk.PromiseName())
	log.Printf("Pipeline name: %s", sdk.PipelineName())

	cfg, err := ctrlcfg.GetConfig()
	if err != nil {
		log.Fatalf("get kube config: %v", err)
	}
	kube, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("build clientset: %v", err)
	}

	switch sdk.WorkflowType() {
	case "promise":
		handlePromise(kube)
	case "resource":
		handleResource(kube)
	default:
		panic("unknown workflow type " + sdk.WorkflowType())
	}
}

func handlePromise(kube *kubernetes.Clientset) {
	switch sdk.PipelineName() {
	case "provision-postgres-db":
		deployPostgres()
	case "litellm-deploy":
		deployLiteLLM(kube)
	default:
		panic("unknown pipeline name " + sdk.PipelineName())
	}
}

func handleResource(kube *kubernetes.Clientset) {
	res, err := sdk.ReadResourceInput()
	if err != nil {
		panic(err)
	}

	team := mustString(get(res, "spec.team"))
	tier := mustString(get(res, "spec.tier"))
	models := toStringSlice(get(res, "spec.models"))

	log.Printf("Setting up LiteLLM for team %s on tier %s with models %v", team, tier, models)
	setupLiteLLMTeam(kube, tier, team, models)

	tiercfg := tierLimits(tier)
	status := kratix.NewStatus()
	statusValues := map[string]any{
		"apiCredsSecretName": team + "-litellm-key",
		"apiEndpoint":        "http://litellm.default.svc.cluster.local:4000",
		"budget": map[string]any{
			"requests_per_minute": tiercfg.RPM,
			"tokens_per_minute":   tiercfg.TPM,
			"max_budget_usd":      fmt.Sprintf("$%.2f", tiercfg.Budget),
			"duration":            tiercfg.BudgetDuration,
		},
	}

	// Preserve original behavior: write empty status first
	if err := sdk.WriteStatus(status); err != nil {
		log.Fatalf("failed to write status: %v", err)
	}

	if ui := get(res, "spec.ui"); asBool(ui) {
		log.Printf("Setting up OpenWebUI")
		deployOpenWebUI(team, models)
		statusValues["uiEndpoint"] = "http://" + team + "-openwebui.default.svc.cluster.local:8080"
	}

	status.Set("ai", statusValues)
	_ = sdk.WriteStatus(status)
}

func get(res kratix.Resource, path string) any {
	v, err := res.GetValue(path)
	if err != nil {
		panic(err)
	}
	return v
}

func mustString(v any) string {
	s, ok := v.(string)
	if !ok {
		panic("expected string but got different type")
	}
	return s
}

func asBool(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func toStringSlice(input any) []string {
	if input == nil {
		return []string{}
	}
	items, ok := input.([]any)
	if !ok {
		panic("expected []any but got different type")
	}
	out := make([]string, len(items))
	for i, v := range items {
		s, ok := v.(string)
		if !ok {
			panic("expected string but got different type")
		}
		out[i] = s
	}
	return out
}
