# envoy-eds-server

### Intro
Envoy EDS server is a working Envoy Discovery Service implementation. It stores in memory an upstream host list and allows any upstream host to register and reregister.

Internally, envoy-eds-server starts two servers: 
- a gRPC server that is invoked by Envoy to fetch the list of upstream hosts.
- a HTTP REST API server where a new host self-registers and deregisters when graceful shutdown happens.

As a good practice, the upstream host should invoke the registration on the REST API server periodically, like a heartbeat.

### Future plan:
- add expiration for an upstream host so that if the heartbeat call is not received then after a while the upstream host becomes evicted.
