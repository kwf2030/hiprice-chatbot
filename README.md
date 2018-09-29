# hiprice-chatbot

## Build
```
docker build -f Dockerfile -t hiprice-chatbot .

// if you do not want to build yourself, a default image is ready in use
docker pull wf2030/hiprice-chatbot:0.1.0
```

## Run
`docker run -d --name hiprice-chatbot --link mariadb:mariadb --link beanstalk:beanstalk hiprice-chatbot`
