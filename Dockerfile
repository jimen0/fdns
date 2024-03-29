FROM golang:1.17-alpine as builder

RUN apk add --no-cache ca-certificates git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

WORKDIR /src/cmd/fdns

RUN go generate
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a --ldflags '-s -w -extldflags "-static"' -tags netgo -installsuffix netgo -o /fdns


FROM alpine:3

RUN apk --no-cache add ca-certificates

COPY --from=builder /fdns /fdns

ENTRYPOINT ["/fdns"]
