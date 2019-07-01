FROM golang:alpine
MAINTAINER thevelop <thevelop@gmail.com>

ARG SERVICE_NAME="go-whatsapp-rest"

ENV CONFIG_ENV="PROD" \
    CONFIG_FILE_PATH="./configs" \
    CONFIG_LOG_LEVEL="INFO" \
    CONFIG_LOG_SERVICE="$SERVICE_NAME"

RUN mkdir /app
ADD . /app/
WORKDIR /app

RUN apk add --no-cache git
RUN go get -u github.com/theveloped/go-whatsapp-rest
RUN go get -u github.com/Rhymen/go-whatsapp
RUN go build -o main .
RUN chmod 777 stores uploads

EXPOSE 3000
HEALTHCHECK --interval=5s --timeout=3s CMD ["curl", "0.0.0.0:3000/api/health"] || exit 1

VOLUME ["/app/stores","/app/uploads"]
CMD ["./main"]
