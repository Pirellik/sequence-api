FROM golang AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -o ./build/api ./cmd/api/main.go
RUN chmod +x /app/build/api

FROM scratch

COPY --from=builder --chmod=0755 /app/build/api /sequence-api

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /app

CMD ["/sequence-api"]