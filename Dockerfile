FROM golang:1.11-alpine3.8

LABEL maintainer="kwf2030 <kwf2030@163.com>" \
      version=0.1.0

RUN echo http://mirrors.aliyun.com/alpine/v3.8/main > /etc/apk/repositories && \
    echo http://mirrors.aliyun.com/alpine/v3.8/community >> /etc/apk/repositories

RUN apk update && \
    apk add --no-cache git nodejs yarn && \
    mkdir -p $GOPATH/src/golang.org/x $GOPATH/src/go.etcd.io /hiprice/admin

WORKDIR $GOPATH/src/golang.org/x

RUN git clone https://github.com/golang/net.git

WORKDIR $GOPATH/src/go.etcd.io

RUN git clone https://github.com/etcd-io/bbolt.git

RUN go get github.com/kwf2030/hiprice-chatbot

WORKDIR $GOPATH/src/github.com/kwf2030/hiprice-chatbot

RUN go build -ldflags "-w -s" && \
    cp hiprice-chatbot /hiprice/chatbot && \
    cp conf.yaml /hiprice/ && \
    cp -r assets/. /hiprice/ && \
    go clean

WORKDIR $GOPATH/src/github.com/kwf2030/hiprice-chatbot/admin

RUN yarn install && \
    yarn run build && \
    cp -r dist/. /hiprice/admin/

WORKDIR /hiprice

ENTRYPOINT ["./chatbot"]