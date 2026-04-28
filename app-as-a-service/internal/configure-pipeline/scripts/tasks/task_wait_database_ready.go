package tasks

import (
	"fmt"
	"time"

	kratix "github.com/syntasso/kratix-go"
	"github.com/syntasso/kratix-marketplace/app/util"
)

func WaitDatabaseReady(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing waitDatabaseReady...")

	ctx, err := util.ReadAppRequestContext(sdk)
	if err != nil {
		return err
	}

	fmt.Println(
		"wait-db-ready inputs:",
		"dbDriver="+ctx.DBDriver,
		"name="+ctx.Name,
		"namespace="+ctx.Namespace,
		fmt.Sprintf("%s=%t", util.VaultOptInLabelKey, ctx.VaultEnabled),
	)

	if ctx.DBDriver == "" || ctx.DBDriver == "none" {
		fmt.Println("wait-db-ready skipped: no database requested")
		return nil
	}
	if ctx.DBDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", ctx.DBDriver)
	}

	pgIdentity := util.DerivePostgresIdentity(ctx.Name)
	timeoutSeconds := util.GetenvIntOrDefault("POSTGRES_READY_TIMEOUT_SECONDS", 900)
	pollIntervalSeconds := util.GetenvIntOrDefault("POSTGRES_READY_POLL_INTERVAL_SECONDS", 5)
	timeout := time.Duration(timeoutSeconds) * time.Second
	pollInterval := time.Duration(pollIntervalSeconds) * time.Second
	deadline := time.Now().Add(timeout)

	var lastObserved string
	for {
		snapshot, err := util.ReadPostgresReadinessSnapshot(ctx.Namespace, pgIdentity.RequestName)
		if err != nil {
			lastObserved = err.Error()
		} else {
			isReady := snapshot.Reconciled == "True" &&
				snapshot.WorksSucceeded == "True" &&
				snapshot.Host != ""
			if isReady {
				if !ctx.VaultEnabled {
					_ = st.Set("database.host", snapshot.Host)
					fmt.Println("Finished executing waitDatabaseReady.")
					return sdk.WriteStatus(st)
				}

				vaultReady := snapshot.VaultAuthPath != "" &&
					snapshot.VaultRole != "" &&
					snapshot.VaultCredentialsPath != ""
				if vaultReady {
					roleReady, roleErr := util.IsVaultRoleConfigured(snapshot)
					if roleErr != nil {
						lastObserved = roleErr.Error()
					} else if roleReady {
						_ = st.Set("database.host", snapshot.Host)
						_ = st.Set("database.vaultEnabled", true)
						_ = st.Set("database.vaultAuthPath", snapshot.VaultAuthPath)
						_ = st.Set("database.vaultRole", snapshot.VaultRole)
						_ = st.Set("database.vaultCredentialsPath", snapshot.VaultCredentialsPath)
						fmt.Println("Finished executing waitDatabaseReady.")
						return sdk.WriteStatus(st)
					} else {
						lastObserved = fmt.Sprintf(
							"vault auth/db roles not ready: role=%q dbCredsPath=%q",
							snapshot.VaultRole,
							snapshot.VaultCredentialsPath,
						)
					}
					time.Sleep(pollInterval)
					continue
				}
			}

			lastObserved = fmt.Sprintf(
				"host=%q reconciled=%q worksSucceeded=%q vaultAuthPath=%q vaultRole=%q vaultCredentialsPath=%q",
				snapshot.Host,
				snapshot.Reconciled,
				snapshot.WorksSucceeded,
				snapshot.VaultAuthPath,
				snapshot.VaultRole,
				snapshot.VaultCredentialsPath,
			)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for postgresql/%s to be ready. last observed: %s", pgIdentity.RequestName, lastObserved)
		}
		time.Sleep(pollInterval)
	}
}
