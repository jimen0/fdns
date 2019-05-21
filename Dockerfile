FROM golang:1.12.5-alpine as builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src

COPY go.mod ./
RUN go mod download
COPY . .

WORKDIR /src/cmd/fdns
RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a --ldflags '-s -w -extldflags "-static"' -tags netgo -installsuffix netgo -o /fdns


FROM alpine:3.9

RUN apk --no-cache add ca-certificates

COPY --from=builder /fdns /fdns

ENTRYPOINT ["/fdns"]
