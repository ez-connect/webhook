FROM docker.io/oven/bun:alpine

ARG arch=amd64

ENV NODE_ENV=production

WORKDIR /webhook

COPY packages/server/dist/ .
COPY packages/plugins/echo/dist/ .

# RUN apt update; \
# 	apt install -y ca-certificates;

CMD	["bun", "webhook.js", "serve"]
