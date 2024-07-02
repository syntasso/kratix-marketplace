# Github Issues Gate

This image includes two commands that, when used in combination, can be used to
gate the execution of a pipeline based on the approval of a Github issue.

## Pre-requisites

The pipeline requires a Github token to be able to interact with the Github API.
Create a secret with the token:

```
kubectl create secret generic github-token --from-literal=token=<YOUR GITHUB TOKEN>
```

Please follow the [Github
documentation](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
for details on how to generate the token.

## Usage

Usage of this image, you must add two workflows to your Promise.

The first workflow will be responsible for creating the issue that will gate the
execution of the next workflow. The definition of this workflow is as follows:

```yaml
workflows:
  resource:
    configure:
      - apiVersion: platform.kratix.io/v1alpha1
        kind: Pipeline
        metadata:
          name: approval-gate
          namespace: default
        spec:
          containers:
            - image: ghcr.io/syntasso/kratix-marketplace/pipeline-github-sign-off-image:v0.1.0
              name: create-issue
              command: [ "create-issue" ]
              env:
              - name: GITHUB_REPOSITORY
                value: myorg/myrepo
              - name: GITHUB_TOKEN
                valueFrom:
                  secretKeyRef:
                    name: github-token
                    key: token
```

At te end of this workflow, the resource will be updated with a link to the
issue, and inform the user that its pending approval.

The second workflow will be responsible for waiting for the approval, and it is
defined as follows:

```yaml
workflows:
  resource:
    configure:
      - # "create-issue workflow" as defined above
      - apiVersion: platform.kratix.io/v1alpha1
        kind: Pipeline
        metadata:
          name: instance-configure
          namespace: default
        spec:
          containers:
            - image: ghcr.io/syntasso/kratix-marketplace/pipeline-github-sign-off-image:v0.1.0
              name: wait-approval
              command: [ "wait-approval" ]
              env:
              - name: GITHUB_TOKEN
                valueFrom:
                  secretKeyRef:
                    name: github-token
                    key: token
            - image: your-promise-image
              name: instance-configure
```

The `wait-approval` command will block the execution of the pipeline until the
issue is closed. It will then create a file at `/kratix/metadata/approval-state`
with "approved" or "rejected" as the content; it will contain the latter when
the issue is closed as `Not planned`.

The next image should is specific to your promise, and should be responsible for
creating the resource you want to create. It must check the
`/kratix/metadata/approval-state` before creating the resources. You chose to
fail the pipeline on rejections, to not deploy anything but surface the
rejection to the resource status, or to ignore and create the resources.

For an example promise, see the `example-promise.yaml` file in this repository.
