FROM debian:8.8

COPY docker-entrypoint.sh /bin
COPY deploy-listener /var/service

ENV LISTENER_ENVIRONMENT production

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["deploy-listener"]
