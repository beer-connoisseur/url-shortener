FROM golang:1.26.2 AS builder

WORKDIR /app

COPY . .

WORKDIR /app/urlshort
RUN go mod download
RUN make -C /app generate-proto
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/urlshort ./cmd/app

FROM ubuntu:22.04
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /bin/urlshort /app/urlshort

CMD ["./urlshort"]