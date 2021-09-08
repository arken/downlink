#! /bin/bash

if [ "$LITESTREAM_ACCESS_KEY_ID" != "" ]; then
    echo "dbs:" > /etc/litestream.yml
    echo "  - path: /app/downlink.db" >> /etc/litestream.yml
    echo "    replicas:" >> /etc/litestream.yml
    echo "      - url: $LITESTREAM_REPLICA_URL" >> /etc/litestream.yml
    echo "        retention: 4h" >> /etc/litestream.yml
    echo "        sync-interval: 15m" >> /etc/litestream.yml
    litestream restore -if-replica-exists -v -o downlink.db "$LITESTREAM_REPLICA_URL"
    litestream replicate -exec /app/downlink
else 
    /app/downlink
fi

