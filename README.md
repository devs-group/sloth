# Requirements

- sqlite3 `brew install sqlite3`
- air `go install github.com/cosmtrek/air@latest`
- golangci `brew install golangci-lint`
- mailhog ( For testing smtp features )

# Get started

After you have installed the above dependecies,
go into the project directory and simply run `npm --prefix ./frontend run dev` to run the frontend
and `air` to run the backend

# Running on the server

The easiest way to run sloth on the server is to use docker

- You can either use plain docker command

```sh
docker run -d \
  -p 80:80 \
  -p 443:443 \
  -p 8080:8080 \
  -p 9090:9090 \
  --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/devs-group/sloth:latest
```

- or use docker-compose

```yaml
version: "3.8"

services:
  app:
    image: sloth:latest
    ports:
      - "80:80"
      - "443:443"
      - "8080:8080"
      - "9090:9090"
    privileged: true
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./deployment/traefik.toml:/etc/traefik/traefik.toml
      - ./deployment/entrypoint.sh:/entrypoint.sh
      - ./:/go/src/app
      - /go/src/app/bin # prevent to override bin folder inside container
```

## Migrations

To create a migration you should only create a new file with an increasing running number.
For example if goose is installed you can use following command

```sh
goose create my_new_table sql
```

To check project related migrations check following path:

```sh
database/migrations
```

from the root project directory

This will create in your actual folder a new file with a timestamp prefix.
Note: The any file prefix must differ and be in the correct order e.g. 1,2,3.. or timestampMMHHss1...
