# hiprice-chatbot
Chatbot for HiPrice.

## Build Docker Image
```
docker build -f Dockerfile -t hiprice-chatbot .

// if you do not want to build yourself, a default image is ready in use
docker pull wf2030/hiprice-chatbot:0.1.0
```

## Run In Docker
`docker run -d --name hiprice-chatbot -p 6200:6200 --link mariadb:mariadb --link beanstalk:beanstalk --link hiprice-web:hiprice-web hiprice-chatbot`

## HiPrice
HiPrice is a wechat personal bot.

It receives product links and crawls regularly, you will be notified when its price goes up/down.

Currently it supports the following websites:
- taobao.com/tmall.com
- jd.com
- suning.com
- amazon.cn
- vip.com
- jumei.com
- mogujie.com
- kaola.com

The whole project has 4 sub projects:
- hiprice-chatbot
Chat bot for HiPrice. Also contains admin console for wechat login. Requires MySQL/MariaDB and Beanstalk.
- hiprice-dispatcher
Dispatcher for HiPrice. Collects product links and dispatches to runners. Requires MySQL/MariaDB and Beanstalk.
- hiprice-runner
Crawler for HiPrice. Deploy as many runners as you can, they are "distributed". Requires Beanstalk and Chrome/Chromium.
- hiprice-web
Web for HiPrice. Manage your own watched products. Only for convenience, usually you should use a CDN instead.

Sub projects has no dependency with each other, but make sure MySQL/MariaDB and Beanstalk is up.

Here is a docker-compose.yml for convenience:

```
version: '3'

services:
  mariadb:
    image: wf2030/mariadb:10.3
    networks:
      - hpnet
    ports:
      - 3306:3306
    environment:
      MYSQL_ROOT_PASSWORD: root

  beanstalk:
    image: wf2030/beanstalk:1.11
    networks:
      - hpnet
    ports:
      - 11300:11300

  hiprice-dispatcher:
    image: wf2030/hiprice-dispatcher:0.1.0
    networks:
      - hpnet
    depends_on:
      - mariadb
      - beanstalk

  hiprice-web:
    image: wf2030/hiprice-web:0.1.0
    networks:
      - hpnet
    ports:
      - 6100:6100

  hiprice-chatbot:
    image: wf2030/hiprice-chatbot:0.1.0
    networks:
      - hpnet
    ports:
      - 6200:6200
    depends_on:
      - mariadb
      - beanstalk

networks:
  hpnet:
    driver: bridge
```
This docker-compose.yml does not contain hiprice-runner, compiles and runs it manually.