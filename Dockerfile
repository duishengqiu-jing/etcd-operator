# Build the manager binary
FROM golang:1.13 as builder

RUN yum -y update && yum -y install upx

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY cmd/ cmd/
COPY pkg/ pkg/

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn"

# Build
RUN go mod download && \
    go build -mod=readonly -o manager main.go && \
    go build -mod=readonly -o backup cmd/backup/main.go && \
    upx manager backup

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
# FROM gcr.io/distroless/static:nonroot
FROM katanomi/distroless-static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER nonroot:nonroot
ENTRYPOINT ["/manager"]

FROM katanomi/distroless-static:nonroot as backup
WORKDIR /
COPY --from=builder /workspace/backup .
USER nonroot:nonroot
ENTRYPOINT ["/backup"]