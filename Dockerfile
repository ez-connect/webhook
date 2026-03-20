FROM scratch

WORKDIR /app

COPY build/* .

EXPOSE 3000
ENTRYPOINT ["/app/webhook", "serve"]
CMD ["-c", "/app/default.config.yaml"]
