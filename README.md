# Aether
Aether Data Platform.

## Build

```bash
    goreleaser build --clean
```

Single platform

```bash
    export GOOS="linux"
    export GOARCH="amd64"
    goreleaser build --single-target --clean
```
