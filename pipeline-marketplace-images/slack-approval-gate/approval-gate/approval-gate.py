#!/usr/bin/env python3
"""Slack approval gate — a Kratix resource-pipeline stage.

Runs as the last container in a Promise's resource.configure pipeline, after the
resource has been written to /kratix/output. It gates provisioning on a human
decision without failing the pipeline:

  approved -> leave /kratix/output intact, exit 0  (Kratix schedules the resource)
  denied   -> clear /kratix/output, exit 0         (nothing is scheduled)
  pending  -> clear /kratix/output, ask Kratix to retry after RETRY_INTERVAL
  no record-> create the request on the backing service, then behave as pending

The stage holds no Slack or state-store credentials — it talks only to the
backing approval service (see the repo README for the contract).
"""

import datetime
import os
import re
import shutil
import sys

import requests
import kratix_sdk as ks

TIMEOUT = 15  # seconds for HTTP calls to the approval server


def parse_duration(value: str) -> datetime.timedelta:
    """Parse a Go-style duration like '20s', '5m', '1h', '500ms'. Defaults to 20s."""
    match = re.fullmatch(r"\s*(\d+)(ms|s|m|h)?\s*", value or "")
    if not match:
        return datetime.timedelta(seconds=20)
    amount, unit = int(match.group(1)), (match.group(2) or "s")
    return {
        "ms": datetime.timedelta(milliseconds=amount),
        "s": datetime.timedelta(seconds=amount),
        "m": datetime.timedelta(minutes=amount),
        "h": datetime.timedelta(hours=amount),
    }[unit]


def clear_dir(path) -> None:
    """Remove the contents of a directory if it exists (leaves the dir itself)."""
    if not path.exists():
        return
    for child in path.iterdir():
        if child.is_dir() and not child.is_symlink():
            shutil.rmtree(child)
        else:
            child.unlink()


def clear_output() -> None:
    clear_dir(ks.get_output_dir())
    clear_dir(ks.get_metadata_dir())


def write_status(sdk: "ks.KratixSDK", message: str) -> None:
    """Write a status message — skipped when running outside a Kratix pod."""
    if not ks.get_metadata_dir().exists():
        return
    status = ks.Status()
    status.set("message", message)
    sdk.write_status(status)


def request_retry(sdk: "ks.KratixSDK", retry_after: datetime.timedelta) -> None:
    """Ask Kratix to reschedule the pipeline — skipped outside a Kratix pod."""
    if not ks.get_metadata_dir().exists():
        return
    sdk.write_retry_after(retry_after, message="Waiting for approval")


def read_request_details(sdk: "ks.KratixSDK"):
    """Return (object_name, spec) for the requested resource.

    The whole ``spec`` is forwarded to the backing service so the stage stays
    generic — it hardcodes no promise-specific fields. In-cluster these come
    from the resource object; when running locally without /kratix/input we fall
    back to env vars so the stage can be exercised by scripts/test-local.sh.
    """
    try:
        resource = sdk.read_resource_input()
    except FileNotFoundError:
        spec = {}
        requester = os.environ.get("REQUESTER_EMAIL")
        if requester:
            spec["requester"] = requester
        return os.environ["KRATIX_OBJECT_NAME"], spec
    return resource.get_name(), (resource.get_value("spec", default={}) or {})


def main() -> int:
    sdk = ks.KratixSDK()

    server_url = os.environ["APPROVAL_SERVER_URL"].rstrip("/")
    token = os.environ["APPROVAL_SERVER_TOKEN"]
    retry_after = parse_duration(os.environ.get("RETRY_INTERVAL", "20s"))
    headers = {"Authorization": f"Bearer {token}"}

    promise_name = sdk.promise_name()
    object_name, spec = read_request_details(sdk)
    request_id = f"approval:{promise_name}:{object_name}"

    # ── Read the current decision from the approval server ──────────────────────
    try:
        resp = requests.get(f"{server_url}/approvals/{request_id}", headers=headers, timeout=TIMEOUT)
    except requests.RequestException as err:
        print(f"ERROR: approval server unreachable: {err} — will retry", file=sys.stderr)
        clear_output()
        write_status(sdk, "Approval server unreachable")
        request_retry(sdk, retry_after)
        return 0

    if resp.status_code == 200:
        decision = resp.json().get("state", "null")
    elif resp.status_code == 404:
        decision = "null"
    else:
        print(
            f"ERROR: approval server returned HTTP {resp.status_code} for {request_id} "
            "— check APPROVAL_SERVER_URL/APPROVAL_SERVER_TOKEN. Will retry.",
            file=sys.stderr,
        )
        clear_output()
        write_status(sdk, f"Approval server error (http {resp.status_code})")
        request_retry(sdk, retry_after)
        return 0

    print(f"Approval state for {request_id}: {decision} (http {resp.status_code})")

    if decision == "approved":
        print("Approved — continuing pipeline")
        write_status(sdk, "Request approved")
        return 0

    if decision == "denied":
        print("Denied — stopping pipeline")
        clear_output()
        write_status(sdk, "Request denied")
        return 0

    if decision == "pending":
        print("Pending — will retry...")
        clear_output()
        write_status(sdk, "Request pending approval")
        request_retry(sdk, retry_after)
        return 0

    # No record yet — create the request (idempotent on the server side).
    print("No approval record — creating one")
    body = {
        "promiseName": promise_name,
        "objectName": object_name,
        "spec": spec,
    }
    try:
        created = requests.post(f"{server_url}/approvals", headers=headers, json=body, timeout=TIMEOUT)
        print(f"Create response: {created.status_code} {created.text}")
    except requests.RequestException as err:
        print(f"WARNING: create request failed: {err} — will retry", file=sys.stderr)

    clear_output()
    write_status(sdk, "Request pending approval")
    request_retry(sdk, retry_after)
    return 0


if __name__ == "__main__":
    sys.exit(main())
