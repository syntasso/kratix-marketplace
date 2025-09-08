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

const (
	appName          = "litellm"
	image            = "ghcr.io/berriai/litellm:v1.75.8-stable"
	httpPort         = 4000
	postgresSecret   = "litellm.litellm-litellm-postgresql.credentials.postgresql.acid.zalan.do"
	postgresHost     = "litellm-litellm-postgresql.default.svc.cluster.local"
	postgresDB       = "litellm"
	pollInterval     = 5 * time.Second
	pollTimeout      = 5 * time.Minute
	dbURLKey         = "DATABASE_URL"
	configSecretName = appName + "-creds"
	configKey        = "config.yaml"
)

func deployPostgres() {
	sdk.WriteOutput("postgres-requests.yaml", []byte(`apiVersion: marketplace.kratix.io/v1alpha2
kind: postgresql
metadata:
  name: `+appName+`
  namespace: default
spec:
  env: dev
  teamId: `+appName+`
  dbName: `+appName+`
`))
}

func deployLiteLLM(kube *kubernetes.Clientset) {
	ctx := context.Background()

	pg, err := waitForSecret(ctx, kube, ns, postgresSecret)
	if err != nil {
		log.Fatalf("waiting for Postgres secret: %v", err)
	}

	user := string(pg.Data["username"])
	pass := url.PathEscape(string(pg.Data["password"]))
	dbURL := fmt.Sprintf("postgresql://%s:%s@%s:5432/%s", user, pass, postgresHost, postgresDB)

	writeDBSecret(ns, appName, dbURL)
	writeDeployment(ns, appName)
	writeService(ns, appName)
}

func waitForSecret(ctx context.Context, kube *kubernetes.Clientset, namespace, name string) (*corev1.Secret, error) {
	var out *corev1.Secret
	err := wait.PollImmediate(pollInterval, pollTimeout, func() (bool, error) {
		s, err := kube.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		out = s
		return true, nil
	})
	return out, err
}

func writeDBSecret(namespace, app, dbURL string) {
	sec := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Secret"},
		ObjectMeta: metav1.ObjectMeta{Name: app + "-db", Namespace: namespace},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{dbURLKey: dbURL},
	}
	writeYAML("20-"+app+"-secret.yaml", sec)
}

func writeDeployment(namespace, app string) {
	replicas := int32(1)
	labels := map[string]string{"app": app}

	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      app,
			Namespace: namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  app,
						Image: image,
						Args:  []string{"--config", "/app/proxy_server_config.yaml"},
						Ports: []corev1.ContainerPort{{ContainerPort: httpPort}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "config-volume",
							MountPath: "/app/proxy_server_config.yaml",
							SubPath:   configKey,
						}},
						EnvFrom: []corev1.EnvFromSource{
							{SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: configSecretName},
							}},
							{SecretRef: &corev1.SecretEnvSource{
								LocalObjectReference: corev1.LocalObjectReference{Name: app + "-db"},
							}},
						},
					}},
					Volumes: []corev1.Volume{{
						Name: "config-volume",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: configSecretName,
								Items:      []corev1.KeyToPath{{Key: configKey, Path: configKey}},
							},
						},
					}},
				},
			},
		},
	}
	writeYAML("30-"+app+"-deployment.yaml", dep)
}

func writeService(namespace, app string) {
	svc := &corev1.Service{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Service"},
		ObjectMeta: metav1.ObjectMeta{Name: app, Namespace: namespace},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": app},
			Type:     corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Protocol:   corev1.ProtocolTCP,
				Port:       httpPort,
				TargetPort: intstr.FromInt(httpPort),
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
