# sql

This Promise provides sql-as-a-Service. The Promise has 1 field `.spec.size`
which can be `small` or `large`.

## Prerequisites

To use this Promise, you will need to create one more secret which is the GCP Service Account permissions needed to create Cloud SQL databases. You can follow Google documentation, or review the following commands via the gcloud CLI:

```bash
export DIR_GCLOUD_CONFIG="${HOME}/.config/gcloud"
export GCP_SERVICE_ACCOUNT_NAME="kratix-sql-$(whoami)-14635" # TOOD ${RANDOM}
export PROJECT_ID=$(sed -n -e 's/^project = //p' ${DIR_GCLOUD_CONFIG}/configurations/config_$(cat ${DIR_GCLOUD_CONFIG}/active_config))

gcloud iam service-accounts create ${GCP_SERVICE_ACCOUNT_NAME} --display-name="Kratix Cloud SQL Manager"
gcloud projects add-iam-policy-binding ${PROJECT_ID} --member="serviceAccount:${GCP_SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" --role="roles/cloudsql.admin"
gcloud iam service-accounts keys create ${DIR_GCLOUD_CONFIG}/kratix-sql-manager-key.json --iam-account "${GCP_SERVICE_ACCOUNT_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
```

Once you have the key json stored in `${DIR_GCLOUD_CONFIG}/kratix-sql-manager-key.json` you can use the following command to generate the Kubernetes secret:

```bash
export DIR_GCLOUD_CONFIG="${HOME}/.config/gcloud"
export PROJECT_ID=$(sed -n -e 's/^project = //p' ${DIR_GCLOUD_CONFIG}/configurations/config_$(cat ${DIR_GCLOUD_CONFIG}/active_config))
cat << EOF | kubectl apply -f -
apiVersion: v1
data:
  credentialsjson: $(cat ${DIR_GCLOUD_CONFIG}/kratix-sql-manager-key.json | base64 -w0)
  project_id: $(echo ${PROJECT_ID} | base64 -w0)
kind: Secret
metadata:
  name: gcp-credentials
EOF
```

## Install Promise

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/sql/gcp/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/sql/gcp/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
