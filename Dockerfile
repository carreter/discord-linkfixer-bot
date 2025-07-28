FROM golang:1.24.4

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /discord-linkfixer-bot

ENTRYPOINT ["/discord-linkfixer-bot"]

