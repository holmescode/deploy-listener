#!/bin/bash
set -e

if [ "$1" = 'listen' ]; then
    exec /deploy-listener "$@"
fi

exec "$@"
