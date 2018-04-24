FROM golang:alpine

RUN apk update && apk add git

RUN apk add --update curl && \
    rm -rf /var/cache/apk/*

COPY . /go/src/app/

WORKDIR /go/src/app

RUN go get -d -v
RUN go install -v

EXPOSE 9090

CMD ["app"]
