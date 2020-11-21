FROM golang:1.15
ARG ES_URL
ENV ES_URL=$ES_URL
WORKDIR /
COPY . .
EXPOSE 8080/tcp
RUN go build
CMD go run parking-sensor-api
