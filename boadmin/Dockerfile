FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates tzdata
ENV TZ Asia/Shanghai

WORKDIR /app
COPY bin/boadmin /app/boadmin
COPY boadmin/etc /app/etc

CMD ["./boadmin", "-f", "etc/boadmin-api.yaml"]
