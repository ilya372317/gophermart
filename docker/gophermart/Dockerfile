FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/gophermart

FROM alpine

COPY --from=builder /app/server /server

CMD ["/server"]