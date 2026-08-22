#!/usr/bin/with-contenv bashio
set -e

# Fixed topology: editor via ingress (internal 8099), viewer on a host port,
# data on the persistent /data volume, production logging format.
export DASHI_DATADIR="/data"
export DASHI_ENV_PRODUCTION="true"
export DASHI_SERVER_EDITOR_ENABLED="true"
export DASHI_SERVER_EDITOR_PORT="8099"
export DASHI_SERVER_VIEWER_ENABLED="true"

# User-configurable options.
export DASHI_ENV_LOGLEVEL="$(bashio::config 'log_level')"
export DASHI_SERVER_VIEWER_PORT="$(bashio::config 'viewer_port')"
if bashio::config.has_value 'viewer_public_url'; then
    export DASHI_SERVER_VIEWER_PUBLICURL="$(bashio::config 'viewer_public_url')"
fi

bashio::log.info "Starting Dashi: editor via ingress (:8099), viewer on :${DASHI_SERVER_VIEWER_PORT}"
exec /usr/bin/dashi start
