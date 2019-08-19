FROM golang:alpine as builder

RUN mkdir /youpod -p

ADD . /youpod

WORKDIR /youpod/cmd

RUN go build -mod=vendor -o main .

FROM alpine:edge

RUN apk add --no-cache ffmpeg
RUN apk add --no-cache youtube-dl=2019.08.13-r0
RUN apk add --no-cache ca-certificates

RUN adduser -S -D -H -h /app appuser
USER appuser
COPY --from=builder /youpod/cmd/main /app/

WORKDIR /app

CMD ["./main"]