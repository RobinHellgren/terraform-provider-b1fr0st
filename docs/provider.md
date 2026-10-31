# Provider: b1fr0st

Configure the provider with an API token from Cloud → Settings → API access.

```terraform
provider "b1fr0st" {
  # Optional; defaults to https://cloud.b1fr0st.dev/api/v1
  api_url = "https://cloud.b1fr0st.dev/api/v1"
}
```

Set `B1FR0ST_API_TOKEN` in the environment. `api_token` may also be provided in configuration, but environment configuration avoids storing the credential in Terraform configuration.
