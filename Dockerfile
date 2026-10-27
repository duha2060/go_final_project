FROM golang:1.26 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /todo .

FROM ubuntu:24.04

WORKDIR /app
COPY --from=build /todo ./todo
COPY web ./web
RUN mkdir -p /app/data

ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/data/scheduler.db

EXPOSE 7540

CMD ["./todo"]
