#!/bin/bash
set -e

if [ "$1" = 'deploy-listener' ]; then
    exec /var/service/deploy-listener "$@"
fi

exec "$@"
