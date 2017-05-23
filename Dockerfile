FROM debian:8.8

WORKDIR /var/service

COPY docker-entrypoint.sh /bin
COPY deploy-listener .

ENV LISTENER_ENVIRONMENT production

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["deploy-listener"]
