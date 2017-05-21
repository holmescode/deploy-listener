FROM debian:8.8

COPY docker-entrypoint.sh /
COPY deploy-listener /

ENV LISTENER_ENVIRONMENT production

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["listen"]
