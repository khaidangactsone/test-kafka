# Sử dụng hình ảnh GoLang chính thức làm hình ảnh cơ sở
FROM golang:1.17 as builder

# Thiết lập thư mục làm việc trong container
WORKDIR /app

# Sao chép mã nguồn vào container
COPY . .

# Cài đặt các phụ thuộc (nếu ứng dụng của bạn có bất kỳ)
RUN go mod download

# Biên dịch ứng dụng Go
RUN go build -o main .

# Sử dụng hình ảnh nhỏ gọn để chạy
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy tệp biên dịch từ bước builder vào thư mục làm việc
COPY --from=builder /app/main .

# Mở cổng 8080
EXPOSE 8090

# Chạy ứng dụng
CMD ["./main"]
