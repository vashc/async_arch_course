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

RUN go build -o $WD/schema_registry ./services/schema_registry

FROM alpine:3.13 AS worker

RUN apk --no-cache add ca-certificates

COPY --from=builder /tmp/bin/schema_registry /app/services/schema_registry

EXPOSE 8010

ENTRYPOINT ["/app/services/schema_registry"]
