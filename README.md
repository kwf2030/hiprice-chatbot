# hiprice-chatbot
Chatbot for HiPrice.

## Build Docker Image
```
docker build -f Dockerfile -t hiprice-chatbot .

// if you do not want to build yourself, a default image is ready in use
docker pull wf2030/hiprice-chatbot:0.1.0
```

## Run In Docker
`docker run -d --name hiprice-chatbot -p 6200:6200 --link mariadb:mariadb --link beanstalk:beanstalk hiprice-chatbot`

## HiPrice
HiPrice is a wechat bot for personal account.

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
- [hiprice-chatbot](https://github.com/kwf2030/hiprice-chatbot)
Chat bot for HiPrice. Also contains admin console for bot login. Requires MySQL/MariaDB and Beanstalk.
- [hiprice-dispatcher](https://github.com/kwf2030/hiprice-dispatcher)
Dispatcher for HiPrice. Collects product links and dispatches to runners. Requires MySQL/MariaDB and Beanstalk.
- [hiprice-runner](https://github.com/kwf2030/hiprice-runner)
Crawler for HiPrice. Deploy as many runners as you can, they are "distributed". Requires Beanstalk and Chrome/Chromium.
- [hiprice-web](https://github.com/kwf2030/hiprice-web)
Web for HiPrice. Manage your own watched products. This project is only for convenience, usually you should use a CDN instead.

Sub projects has no dependency with each other, but make sure MySQL/MariaDB and Beanstalk is up.

You may need the [docker-compose.yml](docker-compose.yml) that lauches all services in one step, the docker-compose.yml does not contain hiprice-runner, compiles and runs it manually.