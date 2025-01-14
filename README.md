# Sloth 🦥

## Overview

**Sloth** is an open-source platform that simplifies container application deployment. Users can configure and deploy
containers through an intuitive web interface. The platform is designed for extensibility and customization, working on
any server or computer.

---

## Features 🚀

- **Web Interface:** User-friendly dashboard for managing container applications.
- **Container Deployment:** Support for container specifications, including ports, URLs, and persistent storage.
- **Automated Routing:** Integrated reverse proxy with Traefik.
- **Cross-Platform Deployment:** Compatible with various system architectures.
- **Security:** OAuth2 authentication with GitHub login.

---

## Tech Stack 🛠️

- **Backend:** Golang
- **Container Technology:** Docker Compose, Traefik
- **Frontend:** Nuxt.js
- **Database:** SQLite

---

## Installation 💻

### Configuration

1. Copy the [.env.example](.env.example)  into the same directory and name it `.env`
    - To be able to use the social logins you need to create the secrets for GitHub and/or Google
    - Github:
        - If you have an organisation visit: https://github.com/organizations/<your-organisation>/settings/applications
        - If you are an individual visit: https://github.com/settings/developers
    - Google:
        - Create a new project in Google Cloud
        - Visit: https://console.cloud.google.com/auth/clients

### Requirements

- **sqlite3:** `brew install sqlite3`
- **golang** (1.21.1 and higher): `brew install go`
- **air:** `go install github.com/cosmtrek/air@latest`
- **golangci-lint:** `brew install golangci-lint`
- **mailhog:** (For testing SMTP features)

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/devs-group/sloth.git
   cd sloth
   ```
2. Start the frontend:
   ```bash
   npm --prefix ./frontend run dev

   OR

   cd frontend
   npm run dev
   ```
3. Start the backend:
   ```bash
   air
   ```

4. Access the web interface at http://localhost:9090 (or `http://localhost:3000`)
    - During development, we proxy requests from http://localhost:9090/_/ to http://localhost:3000/_/ so make sure to
      run the frontend
   > The frontend must run, otherwise you will have errors visiting http://localhost:9090

---

## Running on the Server 🖥️

The easiest way to run Sloth on a server is using Docker.

### Using Docker Command

```bash
docker run -d \
  -p 80:80 \
  -p 443:443 \
  -p 9090:9090 \
  --privileged \
  -v /var/run/docker.sock:/var/run/docker.sock \
  ghcr.io/devs-group/sloth:latest
```

### Using Docker Compose

```yaml
version: "3.8"

services:
  app:
    image: sloth:latest
    ports:
      - "80:80"
      - "443:443"
      - "9090:9090"
    privileged: true
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - ./deployment/traefik.toml:/etc/traefik/traefik.toml
      - ./deployment/entrypoint.sh:/entrypoint.sh
      - ./:/go/src/app
      - /go/src/app/bin  # Prevent overwriting bin folder inside container
```

---

## Migrations ⚙️

To create a migration, simply create a new file with an incrementing number. If `goose` is installed, use:

```bash
goose create my_new_table sql
```

Project-related migrations can be found in:

```
database/migrations
```

from the root project directory.

This will create a new file with a timestamp prefix. Ensure that all file prefixes are unique and ordered correctly (
e.g., `1`, `2`, `3`, or `timestampMMHHss1`).

---

## Deployment

> We use Docker in Docker on production so make sure to mount your local docker.sock into the container to test it

> In case your have issue with the quotes -> "" <- from the .env remove them

1. Test your build locally by running `docker build -f ./deployment/Dockerfile -t sloth/app:latest .` in the **root**
   directory
    - Make sure to change `RUN npm run generate:prod` to `RUN npm run generate` in
      the [release.yml](.github/workflows/release.yml) otherwise you will be redirected to the production page
2. Then you can run:
   `docker run -t -i --env-file .env -v /var/run/docker.sock:/var/run/docker.sock -p 9090:9090 --rm sloth/app:latest`

---

## Tests

> Make sure you are in the [root](.) directory

1. You can run all tests with `go test ./...` locally
    - The [.env.test](backend/tests/.env.test) file can be used to define settings during tests

---

## Contributing 🤝

Contributions are welcome! Please submit a pull request or create an issue on the GitHub repository.

**Enjoy using Sloth! 🦥**

