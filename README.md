# Telegram bot for sending notifications about selected matches for volleyball

Build docker container:
```
export CGO_ENABLED=0
go build
docker build -t bet-bot-go-app .
```

Run docker container:
```
docker run --network host bet-bot-go-app:latest
 ```
