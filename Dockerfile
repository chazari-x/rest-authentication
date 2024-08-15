FROM golang:1.22 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/* cmd/
COPY config/* config/
COPY docs/* docs/
COPY email/* email/
COPY model/* model/
COPY security/* security/
COPY server/* server/
COPY storage/* storage/
COPY main.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.10

RUN adduser -DH server

WORKDIR /app

COPY --from=builder /app/main /app/

RUN chown server:server /app
RUN chmod +x /app

USER server

CMD ["/app/main"]