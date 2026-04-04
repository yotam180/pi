FROM golang:1.26.1-alpine

RUN apk add --no-cache bash git

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

CMD ["go", "test", "./...", "-count=1", "-timeout=300s"]
