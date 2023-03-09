FROM docker.io/golang:1.18 as builder

WORKDIR /app
COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download
COPY ./*.go ./
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main .
RUN chmod +x ./main

FROM docker.io/alpine:3.17.2 AS certificates
RUN apk --no-cache add ca-certificates

FROM scratch
COPY --from=certificates /etc/ssl
COPY --from=builder ./app/main /main
ENTRYPOINT ["/main"]
