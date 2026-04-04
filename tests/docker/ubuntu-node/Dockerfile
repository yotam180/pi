FROM golang:1.26.1-bookworm

RUN apt-get update && apt-get install -y bash git curl && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g tsx

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

CMD ["go", "test", "./...", "-count=1", "-timeout=300s"]
