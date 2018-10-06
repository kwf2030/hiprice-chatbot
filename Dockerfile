FROM golang:1.11-alpine3.8

LABEL maintainer="kwf2030 <kwf2030@163.com>" \
      version=1.0.1

RUN echo http://mirrors.aliyun.com/alpine/v3.8/main > /etc/apk/repositories && \
    echo http://mirrors.aliyun.com/alpine/v3.8/community >> /etc/apk/repositories

RUN apk update && \
    apk add --no-cache git nodejs yarn && \
    mkdir -p /hiprice/bin/admin

WORKDIR /hiprice

RUN git clone https://github.com/kwf2030/hiprice-chatbot.git src

WORKDIR /hiprice/src

RUN git checkout -b b1.0.1 v1.0.1 && \
    go build -ldflags "-w -s" && \
    cp hiprice-chatbot ../bin/chatbot && \
    cp conf.yaml ../bin/ && \
    cp -r assets/. ../bin/ && \
    go clean

WORKDIR /hiprice/src/admin

RUN yarn install && \
    yarn run build && \
    cp -r dist/. ../../bin/admin/

WORKDIR /hiprice/bin

ENTRYPOINT ["./chatbot"]