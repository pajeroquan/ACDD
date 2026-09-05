FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/api /app/api
COPY configs/config.example.yaml /app/configs/config.yaml
ENV TZ=Asia/Shanghai
EXPOSE 8080
ENTRYPOINT ["/app/api", "-config", "/app/configs/config.yaml"]
