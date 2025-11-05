FROM golang:1.24-alpine

WORKDIR /app

# Копируем только go.mod
COPY go.mod .

# Генерируем go.sum внутри контейнера
RUN go mod tidy

# Копируем остальные файлы
COPY . .

RUN go build -o main ./cmd/main.go

CMD ["./main"]