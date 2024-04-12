FROM golang:1.20.4

WORKDIR /app

COPY . .

RUN go build -o main .

CMD ["./main"]