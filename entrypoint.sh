#! /bin/bash

if [ "$LITESTREAM_ACCESS_KEY_ID" != "" ]; then
    litestream restore -if-replica-exists -o downlink.db "$LITESTREAM_REPLICA_URL"
    litestream replicate -exec /app/downlink downlink.db "$LITESTREAM_REPLICA_URL"
else 
    /app/downlink
fi

