package tasks

import (
	"fmt"
	"time"

	kratix "github.com/syntasso/kratix-go"

	"github.com/syntasso/kratix-marketplace/app/util"
)

func WaitDbReady(sdk *kratix.KratixSDK, st kratix.Status) error {
	fmt.Println("Executing WaitDbReady...")
	res, err := sdk.ReadResourceInput()
	if err != nil {
		return fmt.Errorf("read input: %w", err)
	}

	dbDriver := util.MustStringOrEmpty(util.Get(res, "spec.dbDriver"))
	name := util.MustString(util.Get(res, "metadata.name"))
	namespace := util.MustString(util.Get(res, "metadata.namespace"))
	fmt.Println("wait-db-ready inputs:", "dbDriver="+dbDriver, "name="+name, "namespace="+namespace)

	if dbDriver == "" || dbDriver == "none" {
		fmt.Println("wait-db-ready skipped: no database requested")
		return nil
	}
	if dbDriver != "postgresql" {
		return fmt.Errorf("unsupported db driver %q. supported: postgresql", dbDriver)
	}

	pgIdentity := derivePostgresIdentity(name)
	timeoutSeconds := util.GetenvIntOrDefault("POSTGRES_READY_TIMEOUT_SECONDS", 900)
	pollIntervalSeconds := util.GetenvIntOrDefault("POSTGRES_READY_POLL_INTERVAL_SECONDS", 5)
	timeout := time.Duration(timeoutSeconds) * time.Second
	pollInterval := time.Duration(pollIntervalSeconds) * time.Second
	deadline := time.Now().Add(timeout)

	var lastObserved string
	for {
		snapshot, err := readPostgresReadinessSnapshot(namespace, pgIdentity.requestName)
		if err != nil {
			lastObserved = err.Error()
		} else {
			isReady := snapshot.Reconciled == "True" &&
				snapshot.WorksSucceeded == "True" &&
				snapshot.Host != ""
			if isReady {
				_ = st.Set("database.host", snapshot.Host)
				fmt.Println("Finished executing WaitDbReady.")
				return sdk.WriteStatus(st)
			}

			lastObserved = fmt.Sprintf("host=%q reconciled=%q worksSucceeded=%q", snapshot.Host, snapshot.Reconciled, snapshot.WorksSucceeded)
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for postgresql/%s to be ready. last observed: %s", pgIdentity.requestName, lastObserved)
		}
		time.Sleep(pollInterval)
	}
}
