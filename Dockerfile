FROM registry.access.redhat.com/ubi9/go-toolset:latest AS builder
COPY . .
RUN go build -mod=vendor -o .

FROM registry.access.redhat.com/ubi9-micro:latest
COPY --from=builder  /opt/app-root/src/webterminal-proxy /
ENTRYPOINT ["/webterminal-proxy"]
