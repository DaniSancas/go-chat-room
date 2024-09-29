
# Stage 1: Build the Go server
FROM golang:alpine AS build
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/api/main.go

# Stage 2: Run the Go server
FROM alpine
WORKDIR /app
COPY --from=build /app/server .
EXPOSE 8080
CMD ["./server"]
