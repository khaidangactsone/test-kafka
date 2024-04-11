# Sử dụng image base là golang:latest
FROM golang:latest

# Set working directory trong container là /app
WORKDIR /app

# Copy mã nguồn vào thư mục /app trong container
COPY . .

# Biên dịch ứng dụng Go
RUN go build -o main .

# Chạy ứng dụng khi container được khởi động
CMD ["./main"]