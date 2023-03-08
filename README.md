# Web Terminal Proxy Sidecar

A generic websocket proxy for [Web Terminal Operator](https://github.com/redhat-developer/web-terminal-operator) and and [Dev Workspace Operator](https://github.com/devfile/devworkspace-operator).

## Why do we need yet another proxy?

When we want to utilize the Kubernetes `exec` endpoint to execute commands in the pod, we have to authorize using the `Authorization: Bearer` header. JavaScript [WebSocket API](https://websockets.spec.whatwg.org/#the-websocket-interface), which is supported by modern browsers, does not allow additional headers. We have to have a proxy that will accept input from the frontend, and after adding this header, it will send it to the `exec` endpoint.

## Precommit hooks

This repository uses [pre-commit](https://pre-commit.com/). You can install it [here](https://pre-commit.com/#install). To run pre-commit automatically for commits run:

```sh
pre-commit install
```

## Developer guide

TBD

## Deployment guide

TBD
