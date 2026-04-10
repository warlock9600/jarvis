# jarvis

`jarvis` is a daily operations CLI for small engineering tasks: network diagnostics, system helpers, Docker, Kubernetes, and data transforms.

## Build and Run

Requirements:
- Go 1.25 (Docker image: `golang:1.25`)
- `make`

Commands:
- `make build` - build static binary `bin/jarvis`
- `make test` - run unit tests
- `make lint` - run `golangci-lint`
- `make run ARGS="net ip --public"` - run via `go run`
- `make clean` - remove build artifacts

If local Go is unavailable, use Docker:
- `make build-docker`
- `make test-docker`

## CLI shape

`jarvis <domain> <command> [flags]`

Global flags:
- `--json`
- `--no-color`
- `--verbose`, `--debug`, `--quiet`
- `--config`, `--timeout`, `--retries`, `--public-ip-provider`, `--speedtest-bin`, `--kubeconfig`

## Help examples

- `jarvis --help`
- `jarvis net --help`
- `jarvis net ip --help`
- `jarvis help net tls expiry`

## Configuration

Config precedence: CLI flags > env vars > config file.

Default config path:
- `~/.config/jarvis/config.yaml`

Useful env vars:
- `JARVIS_PUBLIC_IP_PROVIDERS`
- `JARVIS_HTTP_TIMEOUT_SECONDS`
- `JARVIS_HTTP_RETRIES`
- `JARVIS_SPEEDTEST_BIN`
- `JARVIS_KUBECONFIG` or `KUBECONFIG`
- `JARVIS_REGISTRY_TOKEN`, `JARVIS_API_TOKEN`

Commands:
- `jarvis config show` (secrets masked)
- `jarvis config path`

## Domains and Examples

### sys
- `jarvis sys password`
- `jarvis sys password --profile strict --length 40 --no-ambiguous`
- `jarvis sys w`
- `jarvis sys hosts take-control`
- `jarvis sys hosts add 10.10.10.10 internal.local`
- `jarvis sys hosts cat`

### net
- `jarvis net ip`
- `jarvis net ip --public --v4`
- `jarvis net speedtest --json`
- `jarvis net dns flush --dry-run`
- `jarvis net check --host example.com --port 443 --http https://example.com`
- `jarvis net dns lookup example.com --type TXT`
- `jarvis net tls expiry example.com:443`

### docker
- `jarvis docker images`
- `jarvis docker exec my-container`
- `jarvis docker prune --dangling --dry-run`

### fs
- `jarvis fs show .`
- `jarvis fs show . --sort size --largest 15 --git-status`
- `jarvis fs show . --tree --depth 2`

### k8s
- `jarvis k8s pods --namespace default --restarts`
- `jarvis k8s images --namespace default`
- `jarvis kube ctx list`
- `jarvis k8s ctx use prod`
- `jarvis k8s ns list`
- `jarvis k8s ns use backend`

### jump
- `jarvis jump bastion`
- `jarvis jump app-prod -L 8080:localhost:8080`
- host autocompletion comes from `~/.ssh/config` and `~/.ssh/config.d/*`

### cat
- `jarvis cat main.go`
- `jarvis cat values.yaml`
- `jarvis cat config.json --style monokai`

### screensaver
- `jarvis screensaver`
- matrix-style terminal animation with clock/date
- exit with `q` or `Ctrl-C`

### data
- `jarvis data b64 encode --input hello`
- `echo aGVsbG8= | jarvis data b64 decode`
- `jarvis data jwt decode <token>`

### completion
- `jarvis completion bash`
- `jarvis completion zsh`
- `jarvis completion fish`

## JSON output contract

When `--json` is set:
- command result is always written to `stdout` as valid JSON;
- logs are written to `stderr` only.

Example payloads:

### `jarvis net ip --json`
```json
{
  "local": [
    {"interface":"en0","version":"ipv4","address":"192.168.1.10"}
  ],
  "public": [
    {"version":"ipv4","address":"203.0.113.11","source":"https://api.ipify.org"}
  ]
}
```

### `jarvis net speedtest --json`
```json
{
  "ping_ms": 15.1,
  "jitter_ms": 3.2,
  "download_mbps": 145.3,
  "upload_mbps": 24.8,
  "raw": {}
}
```

### `jarvis config show --json`
```json
{
  "config_path": "/Users/user/.config/jarvis/config.yaml",
  "secrets": {
    "registry_token": "********",
    "api_token": "********"
  }
}
```

## Return codes

- `0`: success
- `1`: failure
- `2`: partial result (for commands that can return partial data, e.g. `net ip`, `net check`)
