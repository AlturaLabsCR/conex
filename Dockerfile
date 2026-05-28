# syntax=docker/dockerfile:1

FROM node:24-bookworm AS node

FROM golang:1.26.3-bookworm AS build

COPY --from=node /usr/local /usr/local

WORKDIR /src

COPY package.json package-lock.json ./
RUN npm ci

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/conex .

FROM alpine:3.22 AS certs
RUN apk --no-cache add ca-certificates

FROM scratch

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /out/conex /conex

USER 65532:65532
EXPOSE 3080

ENTRYPOINT ["/conex"]
CMD ["serve"]
