package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	intstr "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"

	"sigs.k8s.io/yaml"
)

func deployPostgres() {
	sdk.WriteOutput("postgres-requests.yaml", []byte(`apiVersion: marketplace.kratix.io/v1alpha2
kind: postgresql
metadata:
  name: litellm
  namespace: default
spec:
  env: dev
  teamId: litellm
  dbName: litellm
`))
}

func deployLiteLLM(kube *kubernetes.Clientset) {
	ctx := context.Background()
	ns := "default"
	app := "litellm"

	// derive DATABASE_URL from existing Secret
	pgSecretName := "litellm.litellm-litellm-postgresql.credentials.postgresql.acid.zalan.do"

	var pg *corev1.Secret
	err := wait.PollImmediate(5*time.Second, 5*time.Minute, func() (bool, error) {
		s, err := kube.CoreV1().Secrets(ns).Get(ctx, pgSecretName, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				// not ready yet, retry
				return false, nil
			}
			// non-retryable error
			return false, err
		}
		pg = s
		return true, nil
	})

	if err != nil {
		log.Fatalf("waiting for Postgres secret: %v", err)
	}
	user := string(pg.Data["username"])
	passRaw := string(pg.Data["password"])
	passEsc := url.PathEscape(passRaw)
	host := "litellm-litellm-postgresql.default.svc.cluster.local"
	db := "litellm"
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:5432/%s", user, passEsc, host, db)

	// generateLiteLLMConfigSecret(ns, app)
	generateLitellmSecret(ns, app, dbURL)
	generateLitellmDeployment(ns, app)
	generateLitellmService(ns, app)

	fmt.Println("Wrote manifests to /kratix/output")
}

func generateLitellmSecret(ns, app, dbURL string) {
	sec := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: app + "-db", Namespace: ns},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{
			"DATABASE_URL": dbURL,
		},
	}
	writeYAML("20-"+app+"-secret.yaml", sec)
}

func generateLitellmDeployment(ns, app string) {
	replicas := int32(1)
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app,
			Namespace: ns,
			Labels:    map[string]string{"app": app},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": app}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": app}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  app,
						Image: "ghcr.io/berriai/litellm:v1.75.8-stable",
						Args:  []string{"--config", "/app/proxy_server_config.yaml"},
						Ports: []corev1.ContainerPort{{ContainerPort: 4000}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config-volume",
							MountPath: "/app/proxy_server_config.yaml",
							SubPath:   "config.yaml",
						}},
						EnvFrom: []corev1.EnvFromSource{
							{
								SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: app + "-creds"}},
							},
							{
								SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: app + "-db"}},
							},
						},
					}},
					Volumes: []corev1.Volume{{
						Name: "config-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: app + "-creds",
								Items: []corev1.KeyToPath{{
									Key:  "config.yaml",
									Path: "config.yaml",
								}},
							},
						},
					}},
				},
			},
		},
	}
	writeYAML("30-"+app+"-deployment.yaml", dep)
}

func generateLitellmService(ns, app string) {
	svc := &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: app, Namespace: ns},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": app},
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       4000,
				TargetPort: intstr.FromInt(4000),
			}},
		},
	}
	writeYAML("40-"+app+"-service.yaml", svc)
}

func writeYAML(file string, obj any) {
	b, err := yaml.Marshal(obj)
	if err != nil {
		log.Fatalf("marshal %s: %v", file, err)
	}
	sdk.WriteOutput(file, b)
}

// func getenvDefault(k, def string) string {
// 	if v, ok := os.LookupEnv(k); ok && v != "" {
// 		return v
// 	}
// 	return def
// }
