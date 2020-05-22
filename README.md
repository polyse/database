# polySE: Database
[![Go Report Card](https://goreportcard.com/badge/github.com/polyse/database)](https://goreportcard.com/report/github.com/polyse/database)
![Docker Cloud Build Status](https://img.shields.io/docker/cloud/build/polyse/database)

[ElasticSearch](https://www.elastic.co/) conceptual searchable indexable database.

## Docker image

`docker pull polyse/database:latest`

## Installing

Quick start database:
```bash
git clone https://github.com/polyse/database
cd database
make
```

## Docker

### Start:

Start polySE Database in Docker:

```bash
docker run -d \
    -it \
    -p 9000:9000 \
    --name database \
    --mount type=bind,source=<folder_oh_host>,target=/var/data \
    polyse/database
```

### Environment variables:

`TIMEOUT`

This environment variable is responsible for the timeout for the database response.

Default value: `10ms`.

`LISTEN`

This environment variable is responsible for the network interface that the database is listening on.

Default value in **Docker**: `0.0.0.0:9000`.

`LOG_LEVEL`

This environment variable is responsible for the level of message logging.

Default value: `info`

Available log levels: 
* `debug`
* `info`
* `warn`
* `error`
* `fatal`
* `panic`

`LOG_FMT`

This environment variable sets the log output format.

Default value in **Docker**: `json`

Available formats:
* `console`
* `json`

`DB_FILE`

This environment variable defines the folder in which the database files will be stored.

Default value in **Docker**: `/var/data`

## Documentation

> To see package documentation:
> ```
> go run golang.org/x/tools/cmd/godoc -http=:6060
> ```
> After it you can see documentation in browser by url 
> [http://localhost:6060/pkg/github.com/polyse/database](http://localhost:6060/pkg/github.com/polyse/database)
>
> Url to see documentation to **internal** folder: [http://localhost:6060/pkg/github.com/polyse/database/internal](http://localhost:6060/pkg/github.com/polyse/database/internal)
