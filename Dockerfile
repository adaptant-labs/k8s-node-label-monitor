FROM golang:1.13 as builder

ARG KUBECONFIG
ARG NODE_NAME

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build

FROM scratch
COPY --from=builder /app/k8s-node-label-monitor /app/

ENTRYPOINT [ "/app/k8s-node-label-monitor" ]
