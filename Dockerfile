FROM golang:1.23
WORKDIR /app
COPY go.* .
RUN go mod download
COPY . .
RUN go build -o proxy ./cmd/proxy/proxy.go
EXPOSE 8081/tcp
CMD ["./proxy"]