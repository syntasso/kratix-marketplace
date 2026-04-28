package util

import "regexp"

const (
	defaultLauncherImage      = "ghcr.io/syntasso/kratix-marketplace/app-as-a-service-configure-pipeline:v0.1.0"
	defaultReloaderImage      = "alpine:3.20"
	defaultPostgresWaitImage  = "alpine:3.20"
	VaultOptInLabelKey        = "app-as-a-service.marketplace.kratix.io/vault"
	launcherImageEnv          = "VAULT_LAUNCHER_IMAGE"
	reloaderImageEnv          = "VAULT_RELOADER_IMAGE"
	postgresWaitImageEnv      = "POSTGRES_WAIT_IMAGE"
	launcherInitContainerName = "vault-launcher-init"
	postgresWaitInitName      = "postgres-ready-wait"
	rotationSidecarName       = "vault-rotation-reloader"
	LauncherBinaryPath        = "/vault/bin/vault-env-launcher"
	VaultCredentialsFilePath  = "/vault/secrets/pg-db.env"
	AppPIDFilePath            = "/vault/run/app.pid"
	LauncherVolumeName        = "vault-launcher-bin"
	RuntimeVolumeName         = "vault-runtime"
)

var (
	nonAlnumDashPattern = regexp.MustCompile(`[^a-z0-9-]+`)
	multiDashPattern    = regexp.MustCompile(`-+`)
)
