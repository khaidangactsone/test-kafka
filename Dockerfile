FROM golang:1.18 as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

# Thêm lệnh kiểm tra tệp thực thi
RUN ls -l /root/main

# Thêm quyền thực thi
RUN chmod +x /root/main

EXPOSE 8090

CMD ["./main"]
