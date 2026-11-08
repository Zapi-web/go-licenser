# go-licenser
---
Lightweight, self-hosted software licensing server with Ed25519

![License](https://img.shields.io/badge/License-MIT-blue.svg)
[![Main Pipeline](https://github.com/Zapi-web/go-licenser/actions/workflows/main.yaml/badge.svg)](https://github.com/Zapi-web/go-licenser/actions/workflows/main.yaml)

## Architecture
* **Go** - Core project language, almost everything is built on stdLib
* **Helm** - Flexible Kubernetes manifests for seamless, customizable deployment
* **Github Actions** - CI/CD pipeline that automatically pushes images/charts to the registry
* **GHCR.io** - Container/Chart registry for easy pull
* 
## Features
* **Ed25519-signed licenses:** payloads are signed with a private key and can be verified offline with the public key.
* **Full license lifecycle:** create, activate, renew, revoke, reset HWID and extend expiry.
* **Strict enforcement:** expiry checks, HWID binding and status checks on every operation.
* **Admin API:** management endpoints are protected by a constant-time Bearer token check.
* **Structured logging:** all application logs are emitted with `slog` in JSON format.
* **Lightweight:** highly resource-efficient, consuming significantly less RAM and CPU than production standards.
* **Safety:** distroless Docker base image, non-root user, read-only root filesystem and dropped capabilities.

## Application environment variables

All of these can be overridden through the chart's `env.*` values

| Variable | Description | Default |
| :--- | :--- | :--- |
| `LOG_LEVEL` | Log verbosity (`debug`, `info`, `warn`, `error`) | `info` |
| `VERSION` | License format version embedded in signed payloads | `1` |
| `KEY_GROUPS` | Number of groups in a generated license key | `4` |
| `KEY_GROUPS_LEN` | Characters per key group | `5` |
| `TOKEN_TTL` | Default signed license lifetime | `168h` |
| `CONNECTION_STRING` | PostgreSQL connection URI | *(required)* |
| `PORT` | HTTP listen port (set from `service.port`) | `8080` |

## How To Install

### 1. Create your `values.yaml`

```yaml
image:
  tag: "0.3.0"   # chart/app version from the release tag

secret:
  ED25519_PRIVATE_KEY: "YOUR_BASE64_PKCS8_DER_KEY"
  ADMIN_TOKEN: "YOUR_ADMIN_TOKEN"
  CONNECTION_STRING: "postgres://licenser:password@postgres:5432/licenser?sslmode=disable"

migrations:
  enabled: true
  waitForDb:
    host: postgres

ingress:
  enabled: true
  hosts:
    - host: go-licenser.example.com
      paths:
        - path: /
          pathType: Prefix
```

### 2. Install the chart via OCI registry

```bash
helm install go-licenser oci://ghcr.io/zapi-web/charts/go-licenser \
  --version 0.3.0 \
  -f values.yaml \
  --namespace licenser --create-namespace
```

The chart version and `appVersion` are set from the Git release tag by CI, so use the same version as the release.

### Local development

```bash
# .env with POSTGRES_PASSWORD, ADMIN_TOKEN and ED25519_PRIVATE_KEY
docker compose up --build
```

The API is then available at `http://localhost:8080`.

## API

`/healthz` and `/readyz` are used by Kubernetes probes. 
Public endpoints are used by licensed clients
Admin endpoints require `Authorization: Bearer <ADMIN_TOKEN>`.

| Method | Path | Access | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/healthz` | public | Returns `200 OK` for liveness |
| `GET` | `/readyz` | public | `200 OK`, or `503` when the database is unavailable |
| `POST` | `/v1/activate` | public | Bind a license to a HWID, returns a signed license |
| `POST` | `/v1/renew` | public | Re-issue a signed license for an already activated HWID |
| `POST` | `/v1/licenses` | admin | Create a license (`hwid` and `expires_at` are optional) |
| `GET` | `/v1/licenses/{key}` | admin | Get license details |
| `POST` | `/v1/licenses/{key}/revoke` | admin | Revoke a license |
| `POST` | `/v1/licenses/{key}/reset-hwid` | admin | Unbind a license from its HWID |
| `POST` | `/v1/licenses/{key}/extend` | admin | Extend a license expiry (`expires_at` required) |

Errors are returned as JSON: `{"code": "...", "message": "..."}`

### Example

```bash
# Create a license
curl -s -X POST http://localhost:8080/v1/licenses \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'

# Activate it on a device
curl -s -X POST http://localhost:8080/v1/activate \
  -H "Content-Type: application/json" \
  -d '{"key":"QC279-FMXFA-6XADM-ZAU3P","hwid":"device-fingerprint"}'
```

Activation responses contain a signed license:

```json
{
  "payload": "eyJsaWNlbnNlX2tleSI6...",
  "signature": "..."
}
```

The decoded payload is JSON: `{"license_key": "...", "hwid": "...", "issued_at": "...", "expires_at": "...", "v": 1}`.
