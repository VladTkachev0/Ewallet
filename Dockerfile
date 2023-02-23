FROM golang:1.19.5

RUN mkdir -p /app
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main ./app.go
CMD [ "/app/main" ]