FROM docker.io/golang:1.18 as builder

WORKDIR /app
COPY ./go.mod ./
COPY ./go.sum ./

RUN go mod download
COPY ./*.go ./
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main .
RUN chmod +x ./main

FROM scratch
COPY --from=builder ./app/main /main
ENTRYPOINT ["/main"]
