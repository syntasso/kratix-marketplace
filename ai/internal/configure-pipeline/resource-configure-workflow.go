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

const (
	ns                 = "default"
	masterKeyField     = "LITELLM_MASTER_KEY"
	openWebUIRepoName  = "open-webui"
	openWebUIRepoURL   = "https://open-webui.github.io/helm-charts"
	openWebUIChartRef  = "open-webui/open-webui"
	openWebUIChartVer  = "7.6.0"
	openAIBaseAPI      = "http://litellm.default.svc.cluster.local:4000"
	outputChartPath    = "openwebui.yaml"
	keySecretSuffix    = "-litellm-key"
	litellmCredsSecret = "litellm-creds"
)

func setupLiteLLMTeam(kube *kubernetes.Clientset, tier, team string, models []string) {
	ctx := context.Background()

	sec, err := kube.CoreV1().Secrets(ns).Get(ctx, litellmCredsSecret, metav1.GetOptions{})
	if err != nil {
		log.Fatalf("get secret %q: %v", litellmCredsSecret, err)
	}
	raw, ok := sec.Data[masterKeyField]
	if !ok {
		log.Fatalf("secret %q missing %s", litellmCredsSecret, masterKeyField)
	}
	auth := string(raw)

	secretName := team + keySecretSuffix

	var key string
	sec, err = kube.CoreV1().Secrets(ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		key = generateTeamAndKey(auth, team, tier, models)
	} else {
		key = string(sec.Data[masterKeyField])
	}
	out := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: ns},
		Type:       corev1.SecretTypeOpaque,
		Data:       map[string][]byte{masterKeyField: []byte(key)},
	}
	b, err := yaml.Marshal(out)
	if err != nil {
		log.Fatalf("marshal secret: %v", err)
	}
	if err := sdk.WriteOutput(fmt.Sprintf("%s-litellm-key.yaml", team), b); err != nil {
		log.Fatalf("write output: %v", err)
	}
}

func deployOpenWebUI(team string, models []string) {
	ctx := context.Background()

	settings := helmEnv()
	addOrUpdateRepo(settings, openWebUIRepoName, openWebUIRepoURL)

	actionCfg := new(action.Configuration)
	if err := actionCfg.Init(settings.RESTClientGetter(), ns, "memory", log.Printf); err != nil {
		log.Fatalf("init helm: %v", err)
	}

	inst := action.NewInstall(actionCfg)
	inst.DryRun = true
	inst.ClientOnly = true
	inst.DisableHooks = true
	inst.Replace = true
	inst.Wait = false
	inst.ReleaseName = team + "-openwebui"
	inst.Namespace = ns
	inst.CreateNamespace = false
	inst.ChartPathOptions.Version = openWebUIChartVer

	chartPath, err := inst.ChartPathOptions.LocateChart(openWebUIChartRef, settings)
	if err != nil {
		log.Fatalf("locate chart: %v", err)
	}
	ch, err := loader.Load(chartPath)
	if err != nil {
		log.Fatalf("load chart: %v", err)
	}

	vals := renderValues(team, models)
	rel, err := inst.RunWithContext(ctx, ch, vals)
	if err != nil {
		log.Fatalf("helm render: %v", err)
	}

	// combine manifests
	var buf bytes.Buffer
	buf.WriteString(rel.Manifest)
	for _, hk := range rel.Hooks {
		if buf.Len() > 0 {
			buf.WriteString("\n---\n")
		}
		buf.WriteString(hk.Manifest)
	}

	if err := sdk.WriteOutput(outputChartPath, buf.Bytes()); err != nil {
		log.Fatalf("write output: %v", err)
	}
	fmt.Println("Rendered chart to /kratix/output/" + outputChartPath)
}

func renderValues(team string, models []string) map[string]any {
	yml := `
nameOverride: ` + team + `-openwebui

ollama:
  enabled: false

openaiBaseApiUrl: ` + openAIBaseAPI + `

pipelines:
  enabled: false

extraEnvVars:
  - name: OPENAI_API_KEY
    valueFrom:
      secretKeyRef:
        name: ` + team + keySecretSuffix + `
        key: ` + masterKeyField + `
  - name: DEFAULT_MODELS
    value: ` + strings.Join(models, ",") + `

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
	out := map[string]any{}
	if err := yaml.Unmarshal([]byte(yml), &out); err != nil {
		log.Fatalf("parse values: %v", err)
	}
	return out
}

func helmEnv() *cli.EnvSettings {
	s := cli.New()
	cacheDir := filepath.Join(os.TempDir(), "helm-cache")
	repoFile := filepath.Join(os.TempDir(), "helm-repositories.yaml")
	_ = os.MkdirAll(cacheDir, 0o755)
	s.RepositoryCache = cacheDir
	s.RepositoryConfig = repoFile
	return s
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

	cr, err := repo.NewChartRepository(entry, getter.All(settings))
	if err != nil {
		log.Fatalf("new chart repo %s: %v", name, err)
	}
	cr.CachePath = settings.RepositoryCache
	if _, err := cr.DownloadIndexFile(); err != nil {
		log.Fatalf("update index %s: %v", name, err)
	}
}
