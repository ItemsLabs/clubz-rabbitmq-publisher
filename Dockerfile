FROM golang:1.21-alpine

LABEL maintainer="Sergey Shcherbina <sergeu@ultimatefanlive.com>"

WORKDIR $GOPATH/src/github.com/gameon-app-inc/laliga-matchfantasy-rabbitmq-publisher

COPY . .

RUN go get -d -v ./...

RUN go install -v ./...

CMD ["laliga-matchfantasy-rabbitmq-publisher"]