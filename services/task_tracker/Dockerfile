FROM golang:1.19-alpine AS builder

RUN apk add --no-cache git

ENV CGO_ENABLED=0
ENV GO111MODULE on
ENV GOSUMDB off
ENV WD=/tmp/bin

WORKDIR $WD

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o $WD/task_tracker ./services/task_tracker

FROM alpine:3.13 AS worker

RUN apk --no-cache add ca-certificates

COPY --from=builder /tmp/bin/task_tracker /app/services/task_tracker

EXPOSE 8000

ENTRYPOINT ["/app/services/task_tracker"]
