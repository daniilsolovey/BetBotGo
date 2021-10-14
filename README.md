# Telegram bot for sending notifications about selected matches for volleyball

Build app:
```
make build
```

Run tests:
```
make test
```

Build docker container:
```
docker build -t bet-bot-go-app .
```

Run docker container:

If container with database(betbot-postgres-db) has already been started earlier:

```
docker network create betbotgo
docker network connect betbotgo betbot-postgres-db
```

Command for run database:

```
docker run --name betbo': sudo docker run --name betbot-postgres-db -e POSTGRES_PASSWORD=your_pwd -p 5432:5432 postgres
```

Run container with app:

```
docker-compose up -d
```

Run container with app without compose:

```
docker run --network betbotgo -p 8080:8080 --name bet-bot-go-app bet-bot-go-app:latest
 ```
