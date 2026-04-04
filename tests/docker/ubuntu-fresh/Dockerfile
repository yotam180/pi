FROM golang:1.26.1-bookworm

RUN apt-get update && apt-get install -y bash git && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

CMD ["go", "test", "./...", "-count=1", "-timeout=300s"]
