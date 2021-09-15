FROM alpine:edge

RUN apk update && apk add --no-cache\
    bash \
    tzdata \
    ca-certificates \
    && rm -rf /var/cache/apk/*

COPY BetBotGo /bin/app
COPY config.yaml /etc/BetBot.yaml

CMD ["/bin/app", "--config=/etc/BetBot.yaml"]
