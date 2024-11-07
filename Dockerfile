# Используем Go 1.22.4
FROM golang:1.22.4-alpine AS builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/character

# Используем минималистичный образ для запуска
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Копируем собранное приложение
COPY --from=builder /app/main .

COPY --from=builder /app/config .

EXPOSE 44044-44055

# Устанавливаем точку входа
ENTRYPOINT ["./main"]

# Устанавливаем команду по умолчанию
CMD ["--config", "./local.yaml"]