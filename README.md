# redis-health

Can be used to check the health of your Redis instance. It setups health monitoring server.

## Current checks

- Can connect to Redis
- Can run `info` command
- `loading` is `0` - Redis can't serve traffic while loading data from Disk
- `master_sync_in_progress` is `0` - Redis slaves can't serve traffic while getting initial data from its Master node
- `master_link_status` is `up` - Redis instances that have lost connection to their Redis master should not serve data

## Config

- `LISTEN_ADDR` Heath server listening address (default: `:5000`)
- `REDIS_ADDR` Redis IP+Port to connect to (default: `127.0.0.1:6379`)
- `REDIS_PASS` Redis password to connect with (default: "")

## Running with docker

```
docker run -d xtruder/redis-health
```