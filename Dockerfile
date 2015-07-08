FROM alpine:edge

WORKDIR /usr/src/app
COPY bin/ /usr/src/app

RUN apk --update add go git build-base && \
    GOPATH=/ go get github.com/alecthomas/gometalinter && \
    GOPATH=/ gometalinter --install --update && \
    apk del build-base && rm -fr /usr/share/ri

RUN adduser -u 9000 -D app
USER app

CMD ["/usr/src/app/codeclimate-gometalinter"]
