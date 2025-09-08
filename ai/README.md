# AI Promise

This Kratix promise provides AI-as-a-service, powered by
[LiteLLM](https://litellm.ai/). It allows platform users to request and consume
various AI models in a standardized way.

## High-Level Features

- **Model Selection:** Request specific AI models such as GPT-5, Gemini 2.5 Pro, and Claude Opus 4. Any model that supports the OpenAI API or Ollama API can be integrated.
- **Tiered Deployments:** Choose from different deployment sizes (small, medium, large) to match your performance and cost requirements.
- **Optional UI:** A user interface can be enabled for easier interaction with the AI models.
- **Team Ownership:** Assign ownership of each AI service instance to a specific team.

## Limitations
This Promise as-is only works in a single-cluster setup, that is to say that the
workflows produced by this Promise need to be deployed to the same cluster where
Kratix is deployed. This is due to the fact that the LiteLLM service is deployed
in the same cluster as Kratix, and the Promise Workflows need to make API calls
to the LiteLLM service. Future versions of this Promise will support
multi-cluster. You will also need to ensure the Postgresql Promise is installed
and configured to deploy to the same cluster.

## How it Works

When a user creates a resource request, the promise does the following:

1. **Team and Key Generation:** A new team is created in LiteLLM, and a unique API key is generated for that team. This key is stored in a Kubernetes secret.
2. **Rate Limiting and Budgeting:** Based on the selected tier, rate limits (requests per minute, tokens per minute) and a budget are assigned to the team.
3. **(Optional) UI Deployment:** If requested, an OpenWebUI instance is deployed and configured to use the team's API key and the LiteLLM service.

## Prerequisites

Before installing the promise, you must create a secret named `litellm-creds` in the `default` namespace. This secret contains the LiteLLM master key and the model list configuration. 

Here is an example of the `litellm-creds` secret:

```yaml
---
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: litellm-creds
stringData:
  config.yaml: |
      model_list:
        - model_name: gpt-5
          litellm_params:
            model: gpt-5
            api_base: <YOUR_API_ENDPOINT>
            api_key: <YOUR_OPENAI_API_KEY>
  LITELLM_MASTER_KEY: "sk-123456789"
  LITELLM_SALT_KEY: "sk-012345678"
```

Add in your models under `model_list` as needed.

## Constraints & Configuration

- **Environment:** This promise is a proof-of-concept and should not be used in production without significant modifications. It is intended to be a starting point for building your own AI promise, tailored to your organization's needs.
- **Models:** The available models are defined in the `promise.yaml` and the `litellm-creds` secret. To use your own custom models, you must:
    1.  Modify the `promise.yaml` to include your model names. For example:
        ```yaml
        spec:
          api:
            spec:
              versions:
                - name: v1alpha1
                  schema:
                    openAPIV3Schema:
                      properties:
                        spec:
                          properties:
                            models:
                              items:
                                enum:
                                  - gpt-5
                                  - gemini-2.5-pro
                                  - claude-opus-4
                                  - your-custom-model
        ```
    2.  Update the `model_list` in the `litellm-creds` secret to define the parameters for your models.

- **Dependencies:** This promise depends on a PostgreSQL database, which is
requested from the [PostgreSQL
Promise](https://github.com/syntasso/promise-postgresql). As noted above, you
will need to modify this Promise to deploy request to the Platform cluster.

---

## Installation

To install:
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/ai/promise.yaml
```

To make a resource request (small by default):
```
kubectl apply -f https://raw.githubusercontent.com/syntasso/kratix-marketplace/main/ai/resource-request.yaml
```

## Development

For development see [README.md](./internal/README.md)

## Questions? Feedback?

We are always looking for ways to improve Kratix and the Marketplace. If you run into issues or have ideas for us, please let us know. Feel free to [open an issue](https://github.com/syntasso/kratix-marketplace/issues/new/choose) or [put time on our calendar](https://www.syntasso.io/contact-us). We'd love to hear from you.
