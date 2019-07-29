FROM golang:1.12.7-alpine3.10 as builder

RUN mkdir -p /src

ADD . /src

WORKDIR /src

RUN apk add --no-cache git \
    && go get -d ./... \
    && apk del git
RUN go build -o main .

FROM alpine

RUN apk add --no-cache ffmpeg
RUN apk add --no-cache youtube-dl

RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /src/main /app/

WORKDIR /app
CMD ["./main"]