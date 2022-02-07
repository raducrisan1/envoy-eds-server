# envoy-eds-server

## Intro

Envoy EDS server is a working Endpoint Discovery Service implementation. It stores in memory an upstream host list and allows any upstream host to register and reregister.

Internally, envoy-eds-server starts two servers:

- a gRPC server that is invoked by Envoy to fetch the list of upstream hosts.
- a HTTP REST API server where a new host self-registers and deregisters when graceful shutdown happens.

Envoy makes use of gRPC client stream and is able to receive notifications once an upstream host is registered or unregistered; it does not need poll to update.

As a good practice, the upstream host should invoke the registration on the REST API server periodically, like a heartbeat. This periodic registration update allows the EDS server to become stateless, avoiding in this way the need of a dynamic configuration persistence. The in-memory store of the upstream host list is a good approach because this simplifies the eds server setup (no need to persist the settings) and ensures the configuration becomes consistent even if the EDS server is restarted.

## Docker usage

The following environment variables must be defined:

- `HTTP_LISTEN_PORT` choose the listen port for the HTTP server - the one where upstream hosts are registered / unregistered
- `GRPC_LISTEN_PORT` choose the listen port for the gRPC server - the one consumed by Envoy.
- `EVICTION_TIMEOUT_IN_SEC` choose the time interval to elapse in order to remove the EDS resource because of not receiving a heartbeat API call from the upstream host. Defaults to 42 seconds. If set to zero then the Endpoints (EDS resources) are no more removed and are kept until the envoy-eds-server is restarted.

To start a docker container from the pre-built image, run:

```bash
docker run --name envoy-eds-server --rm -p 8089:8089 -p 8086:8086 raducrisan/envoy-eds-server
```

or

```bash
docker run --name envoy-eds-server --rm --env EVICTION_TIMEOUT_IN_SEC=10 -p 8089:8089 -p 8086:8086 raducrisan/envoy-eds-server
```
