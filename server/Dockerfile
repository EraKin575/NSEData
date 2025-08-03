FROM ubuntu:latest

RUN apt-get update && apt-get install -y golang-go

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o main cmd/api/main.go

CMD ["./main"]

