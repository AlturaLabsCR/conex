FROM alpine:latest

ENV GOLANG_VERSION=1.25.3 \
    GOROOT=/usr/local/go \
    GOPATH=/go \
    PATH=/usr/local/go/bin:$PATH

RUN apk add --no-cache curl make npm

RUN curl -sL https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz -o /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz

WORKDIR /app

COPY . .

RUN make gen && go build -ldflags='-w -s' -o out

ENTRYPOINT ["./out"]
