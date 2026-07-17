# Director Helm Chart

A Helm chart for deploying the Flow Director service on Kubernetes.

## Installation

### OCI Registry

```bash
helm upgrade --install director oci://ghcr.io/muhlba91/hochschule-burgenland-bswe-director-charts/director
```

## Configuration

The following table lists the configurable parameters of the Director chart and their default values.

### Valkey Configuration

| Parameter                                   | Description                                     | Default    |
| ------------------------------------------- | ----------------------------------------------- | ---------- |
| `config.valkey.host`                        | Valkey host                                     | `""`       |
| `config.valkey.port`                        | Valkey port                                     | `6379`     |
| `config.valkey.password`                    | Valkey password                                 | `""`       |
| `config.valkey.existingSecret`              | Existing secret containing Valkey configuration | `""`       |
| `config.valkey.existingSecretKeys.host`     | Key for Valkey host in existing secret          | `host`     |
| `config.valkey.existingSecretKeys.port`     | Key for Valkey port in existing secret          | `port`     |
| `config.valkey.existingSecretKeys.password` | Key for Valkey password in existing secret      | `password` |

### General Configuration

| Parameter                          | Description                     | Default                                                |
| ---------------------------------- | ------------------------------- | ------------------------------------------------------ |
| `replicaCount`                     | Number of replicas              | `1`                                                    |
| `image.repository`                 | Image repository                | `ghcr.io/muhlba91/hochschule-burgenland-bswe-director` |
| `image.tag`                        | Image tag                       | `.Chart.AppVersion`                                    |
| `extraEnv`                         | Extra environment variables     | `[]`                                                   |
| `service.type`                     | Service type                    | `ClusterIP`                                            |
| `service.port`                     | Service port                    | `80`                                                   |
| `ingress.enabled`                  | Enable ingress                  | `false`                                                |
| `gateway.enabled`                  | Enable Gateway API support      | `false`                                                |
| `startupProbe.initialDelaySeconds` | Startup probe initial delay     | `2`                                                    |
| `startupProbe.periodSeconds`       | Startup probe period            | `5`                                                    |
| `startupProbe.failureThreshold`    | Startup probe failure threshold | `30`                                                   |

## Deployment Examples

### Using a Valkey with an existing secret

```yaml
config:
  valkey:
    existingSecret: "my-valkey-secrets"
    existingSecretKeys:
      host: "REDIS_HOST"
      port: "REDIS_PORT"
      password: "REDIS_PASSWORD"
```

### Adding extra environment variables

```yaml
extraEnv:
  - name: LOG_LEVEL
    value: "debug"
  - name: AUTH_SECRET
    valueFrom:
      secretKeyRef:
        name: director-auth
        key: secret
```
