#!/usr/bin/env python3
"""claim — ref-counted "claims" pipeline for the NamespaceClaim promise.

Dispatches on the Kratix workflow action (KRATIX_WORKFLOW_ACTION via the SDK):
  configure -> ensure  : create the shared provider request if absent
  delete    -> release : delete the provider request iff this is the last live
                         claim for the same target

The provider request is created/deleted imperatively via the Kubernetes API so
it is owned by nothing and survives any single claim's deletion. (Writing it to
/kratix/output would make it Work-owned and it would be pruned with the first
claim -- the bug this whole pattern exists to avoid.)

To make another resource claimable: copy this file, set TARGET_FIELD, and adjust
the PROVIDER_* constants and the manifest built in ensure().
"""
import os

# The shared resource this claim manages: a request to the existing `namespace`
# promise. Namespace-specific; everything else is generic ref-counting.
PROVIDER_GROUP = "marketplace.kratix.io"
PROVIDER_VERSION = "v1alpha1"
PROVIDER_PLURAL = "namespaces"
PROVIDER_KIND = "namespace"
PROVIDER_NAMESPACE = "default"


def target_field():
    """Dot path to the shared-resource identity in the claim spec."""
    return os.getenv("TARGET_FIELD", "spec.namespaceName").lstrip(".")


def _get_path(obj, dotted):
    cur = obj
    for part in dotted.split("."):
        if not isinstance(cur, dict):
            return None
        cur = cur.get(part)
    return cur


def count_live_claims(items, field, target):
    """Pure: count claims whose `field` == target and which are NOT terminating.

    A claim being deleted is still returned by the API but carries a
    metadata.deletionTimestamp; excluding those makes the count reach zero only
    when no live claim remains.
    """
    count = 0
    for item in items:
        if _get_path(item, field) != target:
            continue
        if (item.get("metadata") or {}).get("deletionTimestamp"):
            continue
        count += 1
    return count


def _api():
    from kubernetes import client, config
    try:
        config.load_incluster_config()
    except Exception:
        config.load_kube_config()
    return client.CustomObjectsApi()


def ensure(api, target, cluster_name):
    """Idempotently create the provider request named after `target`."""
    from kubernetes.client.rest import ApiException

    spec = {"namespaceName": target}
    if cluster_name:
        spec["clusterName"] = cluster_name
    body = {
        "apiVersion": f"{PROVIDER_GROUP}/{PROVIDER_VERSION}",
        "kind": PROVIDER_KIND,
        "metadata": {"name": target, "namespace": PROVIDER_NAMESPACE},
        "spec": spec,
    }
    try:
        api.create_namespaced_custom_object(
            group=PROVIDER_GROUP,
            version=PROVIDER_VERSION,
            namespace=PROVIDER_NAMESPACE,
            plural=PROVIDER_PLURAL,
            body=body,
        )
        print(f"claim: created {PROVIDER_KIND}/{target}")
    except ApiException as exc:
        if exc.status == 409:
            print(f"claim: {PROVIDER_KIND}/{target} already exists")
        else:
            raise


def release(api, gvk, plural, target, field):
    """Delete the provider request iff no live sibling claim remains."""
    from kubernetes.client.rest import ApiException

    objs = api.list_cluster_custom_object(
        group=gvk.group, version=gvk.version, plural=plural
    )
    count = count_live_claims(objs.get("items", []), field, target)
    print(f"claim: {count} live claim(s) for {target}")
    if count > 0:
        return
    try:
        # Async delete (no wait): does not block on the provider RR's finalizer
        # cascade. Kratix garbage-collects its Work afterwards.
        api.delete_namespaced_custom_object(
            group=PROVIDER_GROUP,
            version=PROVIDER_VERSION,
            namespace=PROVIDER_NAMESPACE,
            plural=PROVIDER_PLURAL,
            name=target,
        )
        print(f"claim: deleted {PROVIDER_KIND}/{target}")
    except ApiException as exc:
        if exc.status != 404:
            raise


def main():
    import kratix_sdk as ks

    sdk = ks.KratixSDK()
    resource = sdk.read_resource_input()

    field = target_field()
    target = resource.get_value(field, default=None)
    if not target:
        raise SystemExit(f"claim: {field} is required")

    api = _api()

    if sdk.workflow_action() == "delete":
        plural = os.getenv("KRATIX_CRD_PLURAL")
        if not plural:
            raise SystemExit("claim: KRATIX_CRD_PLURAL is not set")
        release(api, resource.get_group_version_kind(), plural, target, field)
    else:
        cluster_name = resource.get_value("spec.clusterName", default="") or ""
        ensure(api, target, cluster_name)


if __name__ == "__main__":
    main()
