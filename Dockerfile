# FROM alpine:edge

# RUN apk update && apk add \
#     bash \
#     ca-certificates \
#     && rm -rf /var/cache/apk/*

# COPY BetBot /bin/app
# COPY config.yaml /etc/BetBot.yaml


# CMD ["/bin/app", "--config=/etc/BetBot.yaml"]
