# Web Terminal Proxy Sidecar

A generic websocket proxy for [Web Terminal Operator](https://github.com/redhat-developer/web-terminal-operator) and and [Dev Workspace Operator](https://github.com/devfile/devworkspace-operator).

## Why do we need yet another proxy?

When we want to utilize the Kubernetes `exec` endpoint to execute commands in the pod, we have to authorize using the `Authorization: Bearer` header. JavaScript [WebSocket API](https://websockets.spec.whatwg.org/#the-websocket-interface), which is supported by modern browsers, does not allow additional headers. We have to have a proxy that will accept input from the frontend, and after adding this header, it will send it to the `exec` endpoint. 
Openshift also does not include the `Access-Control-Allow-Origin` header, so we have to pass requests from the frontend through the proxy to allow frontend to parse the responses.
## Precommit hooks

This repository uses [pre-commit](https://pre-commit.com/). You can install it [here](https://pre-commit.com/#install). To run pre-commit automatically for commits run:

```sh
pre-commit install
```

## Developer guide

### Requirements

Go >= 1.19

### Running locally

To run the application locally, you can run the following command:
```
go run .
```



### Deployment guide

To deploy the application as a sidecar for Backstage deployment, you must create a `Route`, modify the `Service` resource and add a sidecar to the Backstage Deployment. You also have to build the webterminal-proxy image and push it to your registry.

**Route**

```yaml
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: backstage-webterminal
spec:
  port:
    targetPort: 8080
  to:
    kind: Service
    name: backstage-instance
  host: backstage-instance.example.com
  path: "/webterminal"

```

**Service**

```yaml
 # ...Your Backstage service definition...
  ports:
    # ...
    - port: 8080
      targetPort: 8080
    # ...
```

**Deployment**

```yaml
 # ...Your Backstage deployment...
  spec:
    #...
    containers:
        # ...
        - name: webterminal
          image: image-registry.example.com/webterminal-proxy:latest
          command: ["./webterminal-proxy"]
          ports:
            - containerPort: 8080
        # ...

```