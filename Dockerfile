FROM golang:1.22 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 go build -o cyphera-bq .

FROM gcr.io/distroless/static-debian12
COPY --from=build /app/cyphera-bq /cyphera-bq
COPY config/cyphera.yaml /etc/cyphera/cyphera.yaml
EXPOSE 8080
ENTRYPOINT ["/cyphera-bq"]
