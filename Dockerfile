FROM golang:1.15
WORKDIR /
COPY . .
EXPOSE 8080/tcp
RUN go build
CMD go run parking-sensor-api
