## Setup for running Gloo tests locally

### e2e Tests

Instructions for setting up and running the end-to-end tests can be found [here](./e2e#end-to-end-tests).

### Kubernetes e2e Tests

Instructions for setting up and running the kubernetes end-to-end tests can be found [here](./kube2e#kubernetes-end-to-end-tests).

### Consult/Vault e2e Tests

Instructions for setting up and running the consul and vault end-to-end tests can be found [here](./consulvaulte2e).

## Debugging Tests

# Gloo Tests

Some of the gloo tests use a listener on 127.0.0.1 rather than 0.0.0.0 and will only run on linux (e.g. fault injection).

If you’re developing on a mac, ValidateBootstrap will not run properly because it uses the envoy binary for validation mode (which only runs on linux). See rbac_jwt_test.go for an example.