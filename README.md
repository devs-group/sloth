# Sloth 🦥

## Overview

**Sloth** is an open-source platform that simplifies container application deployment. Users can configure and deploy containers through an intuitive web interface. The platform is designed for extensibility and customization, working on any server or computer.

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

### Requirements

- **sqlite3:** `brew install sqlite3`
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
   ```
3. Start the backend:
   ```bash
   air
   ```

4. Access the web interface at `http://localhost:8080`

---

## Running on the Server 🖥️

The easiest way to run Sloth on a server is using Docker.

### Using Docker Command
```bash
docker run -d \
  -p 80:80 \
  -p 443:443 \
  -p 8080:8080 \
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
      - "8080:8080"
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

This will create a new file with a timestamp prefix. Ensure that all file prefixes are unique and ordered correctly (e.g., `1`, `2`, `3`, or `timestampMMHHss1`).

---

## Contributing 🤝

Contributions are welcome! Please submit a pull request or create an issue on the GitHub repository.

**Enjoy using Sloth! 🦥**

