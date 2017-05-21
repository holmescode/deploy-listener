FROM alpine:3.5

WORKDIR /app
COPY deploy-listener .

ENV LISTENER_ENVIRONMENT production

ENTRYPOINT ["/app/deploy-listener"]
