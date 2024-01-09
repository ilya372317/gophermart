FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/gophermart

FROM alpine

RUN mkdir cmd
RUN mkdir cmd/accrual

COPY --from=builder /app/server /server
COPY --from=builder /app/cmd/accrual /cmd/accrual

CMD ["/server"]