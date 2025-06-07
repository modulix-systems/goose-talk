FROM golang:1.24.4-bookworm

WORKDIR /app

COPY go.mod go.sum .

RUN go mod tidy

COPY . .

ENTRYPOINT ["go", "run", "./cmd/app"]
