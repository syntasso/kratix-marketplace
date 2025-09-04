package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

func setupLiteLLMTeam(kube *kubernetes.Clientset, tier, team string, models []string) {
	sec, err := kube.CoreV1().Secrets("default").Get(context.Background(), "litellm-creds", metav1.GetOptions{})
	if err != nil {
		log.Fatalf("failed to get secret: %v", err)
	}

	rawVal, ok := sec.Data["LITELLM_MASTER_KEY"]
	if !ok {
		log.Fatalf("secret missing LITELLM_MASTER_KEY field")
	}

	auth := string(rawVal)

	key := generateTeamAndKey(auth, team, tier, models)

	sec = &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      team + "-litellm-key",
			Namespace: "default",
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"LITELLM_MASTER_KEY": []byte(key),
		},
	}

	b, err := yaml.Marshal(sec)
	if err != nil {
		log.Fatalf("marshal secret: %v", err)
	}

	sdk.WriteOutput(team+"-litellm-key.yaml", b)

}

func deployOpenWebUI(team string, models []string) {
	ctx := context.Background()

	releaseName := team + "-openwebui"
	namespace := "default"
	chartRef := "open-webui/open-webui"
	chartVersion := "7.6.0"

	modelsStr := strings.Join(models, ",")

	valuesYAML := `
nameOverride: ` + team + `-openwebui
ollama:
  enabled: false

openaiBaseApiUrl: http://litellm.default.svc.cluster.local:4000

pipelines:
  enabled: false

extraEnvVars:
  - name: OPENAI_API_KEY
    valueFrom:
      secretKeyRef:
        name: ` + team + `-litellm-key
        key: LITELLM_MASTER_KEY
  - name: DEFAULT_MODELS
    value: ` + modelsStr + `

resources:
  requests:
    cpu: "1000m"
    memory: "900Mi"
  limits:
    cpu: "2000m"
    memory: "2Gi"

service:
  type: ClusterIP

ingress:
  enabled: false
`

	// Helm env
	settings := cli.New()
	cacheDir := filepath.Join(os.TempDir(), "helm-cache")
	repoFile := filepath.Join(os.TempDir(), "helm-repositories.yaml")
	_ = os.MkdirAll(cacheDir, 0o755)
	settings.RepositoryCache = cacheDir
	settings.RepositoryConfig = repoFile

	// Repos
	addOrUpdateRepo(settings, "open-webui", "https://open-webui.github.io/helm-charts")

	// Action config
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), namespace, "memory", log.Printf); err != nil {
		log.Fatalf("init helm action config: %v", err)
	}

	// Install action in dry mode
	inst := action.NewInstall(actionConfig)
	inst.DryRun = true
	inst.ClientOnly = true
	inst.DisableHooks = true
	inst.Replace = true
	inst.Wait = false
	inst.ReleaseName = releaseName
	inst.Namespace = namespace
	inst.CreateNamespace = false
	inst.ChartPathOptions.Version = chartVersion

	chartPath, err := inst.ChartPathOptions.LocateChart(chartRef, settings)
	if err != nil {
		log.Fatalf("locate chart: %v", err)
	}
	ch, err := loader.Load(chartPath)
	if err != nil {
		log.Fatalf("load chart: %v", err)
	}

	vals := map[string]any{}
	if err := yaml.Unmarshal([]byte(valuesYAML), &vals); err != nil {
		log.Fatalf("parse values: %v", err)
	}

	rel, err := inst.RunWithContext(ctx, ch, vals)
	if err != nil {
		log.Fatalf("helm render: %v", err)
	}

	// Single combined YAML
	var buf bytes.Buffer
	buf.WriteString(rel.Manifest)
	for _, hk := range rel.Hooks {
		if buf.Len() > 0 {
			buf.WriteString("\n---\n")
		}
		buf.WriteString(hk.Manifest)
	}

	if err := sdk.WriteOutput("openwebui.yaml", buf.Bytes()); err != nil {
		log.Fatalf("write output: %v", err)
	}

	fmt.Println("Rendered chart to /kratix/output/openwebui.yaml")
}

func addOrUpdateRepo(settings *cli.EnvSettings, name, url string) {
	rf := repo.NewFile()
	if _, err := os.Stat(settings.RepositoryConfig); err == nil {
		if f, err := repo.LoadFile(settings.RepositoryConfig); err == nil {
			rf = f
		}
	}
	entry := &repo.Entry{Name: name, URL: url}
	if existing := rf.Get(name); existing != nil {
		existing.URL = url
	} else {
		rf.Update(entry)
	}
	if err := rf.WriteFile(settings.RepositoryConfig, 0o644); err != nil {
		log.Fatalf("write repo file: %v", err)
	}
	chRepo, err := repo.NewChartRepository(entry, getter.All(settings))
	if err != nil {
		log.Fatalf("new chart repo %s: %v", name, err)
	}
	chRepo.CachePath = settings.RepositoryCache
	if _, err := chRepo.DownloadIndexFile(); err != nil {
		log.Fatalf("update index %s: %v", name, err)
	}
}
