package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	kratix "github.com/syntasso/kratix-go"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func main() {
	sdk := kratix.New()
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalf("usage: %s <pipeline-name>", os.Args[0])
	}
	arg := args[0]
	fmt.Println("Pipeline requested:", arg)
	st, err := sdk.ReadStatus()
	if err != nil {
		st = kratix.NewStatus() // start with empty status
	}
	switch arg {
	case "resource-configure":
		if err := runResource(sdk, st); err != nil {
			log.Fatalf("resource pipeline: %v", err)
		}
	case "database-configure":
		if err := runDatabase(sdk, st); err != nil {
			log.Fatalf("database pipeline: %v", err)
		}
	default:
		log.Fatalf("unknown pipeline %q", sdk.PipelineName())
	}
	fmt.Println("Finished executing main.")
}

func runResource(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing runResource...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	image := mustString(get(res, "spec.image"))
	name := mustString(get(res, "spec.name"))
	namespace := mustString(get(res, "metadata.namespace"))
	servicePort := mustString(get(res, "spec.service.port"))
	fmt.Println("resource-configure inputs:", "image="+image, "name="+name, "namespace="+namespace, "servicePort="+servicePort)

	deployYAML := runKubectl(
		"create", "deployment", name,
		"--namespace="+namespace,
		"--replicas=1",
		"--image="+image,
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("deployment.yaml", deployYAML); err != nil {
		return fmt.Errorf("write deployment: %w", err)
	}

	svcYAML := runKubectl(
		"create", "service", "nodeport", name,
		"--namespace="+namespace,
		"--tcp="+servicePort+":8080",
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("service.yaml", svcYAML); err != nil {
		return fmt.Errorf("write service: %w", err)
	}

	rule := fmt.Sprintf("%s.local.gd/*=%s:%s", name, name, servicePort)
	ingYAML := runKubectl(
		"create", "ingress", name,
		"--namespace="+namespace,
		"--class=nginx",
		"--rule="+rule,
		"--dry-run=client", "-o", "yaml",
	)
	if err := sdk.WriteOutput("ingress.yaml", ingYAML); err != nil {
		return fmt.Errorf("write ingress: %w", err)
	}

	endpoint := fmt.Sprintf("http://%s.local.gd:31338", name)
	_ = st.Set("message", "deployed to "+endpoint)
	_ = st.Set("endpoint", endpoint)
	_ = st.Set("replicas", int64(1))
	fmt.Println("Finished executing runResource.")
	return nil
}

func runDatabase(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing runDatabase...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	dbDriver := mustStringOrEmpty(get(res, "spec.dbDriver"))
	name := mustString(get(res, "metadata.name"))
	namespace := mustString(get(res, "metadata.namespace"))
	fmt.Println("database-configure inputs:", "dbDriver="+dbDriver, "name="+name, "namespace="+namespace)

	if dbDriver == "" || dbDriver == "none" {
		st := kratix.NewStatus()
		_ = st.Set("database", nil) // -> database: null
		return sdk.WriteStatus(st)
	}
	if dbDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", dbDriver)
	}

	teamID := name + "team"
	dbName := name + "db"

	// 1) update existing deployment with env
	deploy, err := readDeployment("/kratix/output/deployment.yaml")
	if err != nil {
		return fmt.Errorf("read existing deployment: %w", err)
	}
	secretRef := fmt.Sprintf("%s.%s-%s-postgresql.credentials.postgresql.acid.zalan.do", teamID, teamID, dbName)

	env := []corev1.EnvVar{
		{
			Name: "PGPASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretRef},
					Key:                  "password",
				},
			},
		},
		{
			Name: "PGUSER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: secretRef},
					Key:                  "username",
				},
			},
		},
		{
			Name:  "PGHOST",
			Value: fmt.Sprintf("%s-%s-postgresql.default.svc.cluster.local", teamID, dbName),
		},
		{
			Name:  "DBNAME",
			Value: dbName,
		},
	}
	deploy.Spec.Template.Spec.Containers[0].Env = env
	if err := writeYAMLObject(sdk, "deployment.yaml", &deploy); err != nil {
		return fmt.Errorf("write updated deployment: %w", err)
	}

	// 2) write postgresql CR using unstructured
	if err := os.MkdirAll(filepath.Clean("/kratix/output/platform"), 0o755); err != nil {
		return fmt.Errorf("mkdir platform: %w", err)
	}
	pg := &unstructured.Unstructured{Object: map[string]any{}}
	pg.SetAPIVersion("marketplace.kratix.io/v1alpha2")
	pg.SetKind("postgresql")
	pg.SetName(dbName)
	pg.SetNamespace(namespace)
	_ = unstructured.SetNestedField(pg.Object, teamID, "spec", "teamId")
	_ = unstructured.SetNestedField(pg.Object, dbName, "spec", "dbName")
	if err := writeYAMLMap(sdk, "platform/postgresql.yaml", pg.Object); err != nil {
		return fmt.Errorf("write postgresql request: %w", err)
	}

	ds := []kratix.DestinationSelector{
		{Directory: "platform", MatchLabels: map[string]string{"environment": "platform"}},
	}
	if err := sdk.WriteDestinationSelectors(ds); err != nil {
		return fmt.Errorf("write destination selectors: %w", err)
	}

	_ = st.Set("database.teamId", teamID)
	_ = st.Set("database.dbName", dbName)
	fmt.Println("Finished executing runDatabase.")
	return sdk.WriteStatus(st)
}

func runKubectl(args ...string) []byte {
	fmt.Printf("Executing runKubectl: %v", args)
	cmd := exec.Command("kubectl", args...)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Fatalf("kubectl %v failed: %v\nstderr:\n%s", args, err, string(ee.Stderr))
		}
		log.Fatalf("kubectl %v failed: %v", args, err)
	}
	return out
}

func writeYAMLObject(sdk *kratix.KratixSDK, filename string, obj any) error {
	b, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	return sdk.WriteOutput(filename, b)
}

func writeYAMLMap(sdk *kratix.KratixSDK, filename string, m map[string]any) error {
	b, err := yaml.Marshal(m) // preserves apiVersion/kind/metadata/spec casing
	if err != nil {
		return err
	}
	return sdk.WriteOutput(filename, b)
}

func readDeployment(path string) (appsv1.Deployment, error) {
	var d appsv1.Deployment
	b, err := os.ReadFile(path)
	if err != nil {
		return d, err
	}
	if err := yaml.Unmarshal(b, &d); err != nil {
		return d, err
	}
	// Ensure TypeMeta present when re-serializing
	if d.APIVersion == "" {
		d.APIVersion = "apps/v1"
	}
	if d.Kind == "" {
		d.Kind = "Deployment"
	}
	return d, nil
}

func get(res kratix.Resource, path string) any {
	v, err := res.GetValue(path)
	if err != nil {
		log.Fatalf("get %s: %v", path, err)
	}
	return v
}

func mustString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int32:
		return strconv.FormatInt(int64(t), 10)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		log.Fatalf("want string got %T", v)
		return ""
	}
}

func mustStringOrEmpty(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	default:
		return mustString(v)
	}
}
