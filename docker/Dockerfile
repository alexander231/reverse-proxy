FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

RUN ls
COPY cmd ./cmd
COPY /config ./config
COPY /pkg ./pkg

RUN CGO_ENABLED=0 go build -o bin/reverse-proxy cmd/main.go

FROM scratch

COPY --from=build /app/bin/reverse-proxy /
COPY --from=build /app/config /config

ENTRYPOINT [ "/reverse-proxy" ]
EXPOSE 8080


