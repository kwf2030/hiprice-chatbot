FROM alpine:3.8

LABEL maintainer="kwf2030 <kwf2030@163.com>" \
      version=1.11

RUN echo http://mirrors.aliyun.com/alpine/v3.8/main > /etc/apk/repositories && \
    echo http://mirrors.aliyun.com/alpine/v3.8/community >> /etc/apk/repositories

RUN apk update && \
    apk upgrade && \
    apk add --no-cache git && \
    apk add --virtual build-base && \
    mkdir -p /beanstalk/bin /beanstalk/data

WORKDIR /beanstalk

RUN git clone https://github.com/kwf2030/beanstalkd.git src

WORKDIR /beanstalk/src

RUN git checkout -b b1.11 v1.11 && \
    make && \
    cp beanstalkd ../bin/ && \
    make clean

ENV PORT=11300

WORKDIR /beanstalk/bin

ENTRYPOINT ["/bin/sh", "-c", "./beanstalkd", "-b", "/beanstalk/data", "-f", "60000", "-l", "0.0.0.0", "-p", "$PORT", "-z", "16777216"]