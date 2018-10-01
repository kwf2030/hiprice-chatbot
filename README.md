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
