# Aether

Aether is a smart data management platform designed to organize and tag data stored in Object Store,
with advanced support for grouping via datasets and automatic tag generation powered by AI agents.
The platform integrates with both S3 and a datastore to:

Efficiently tag and organize data
Enable fast querying and retrieval
Ensure storage optimization by avoiding duplicates and applying intelligent structuring

Aether simplifies data governance by combining automation, scalability, and intelligent metadata handling—making it ideal for teams managing large-scale, unstructured data.

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
