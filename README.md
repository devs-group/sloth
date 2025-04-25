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
- **Security:** OAuth2 authentication with GitHub and Google login.

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

- **Golang** (1.23.3 and higher):
  - For using Goose: `brew install go`
- **Goose** (3.18.0 and higher):
  - `go install github.com/pressly/goose/v3/cmd/goose@v3.18.0`
- **Docker** (27.4.0 and higher):
  - For Mac and Linux we recommend: [Docker Desktop](https://docs.docker.com/desktop/)
  - For Servers (Linux) we recommend: [Docker Engine](https://docs.docker.com/engine/)

### Docker mounts

Docker in Docker is a special topic and we figured it might make sense to define `/var/app` as the work directory for
all the services so we can be sure that we can mount these also on hosts using Mac (We are not counting in Windows
at this time). `/var/folder` is already implemented in Docker Desktop as a known virtual mount. This is explicitly
documented [here](https://docs.docker.com/desktop/settings-and-maintenance/settings/#virtual-file-shares).

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/devs-group/sloth.git
   cd sloth
   ```
2. Create a network called `traefik` on your local machine with: `docker network create traefik`
3. Start everything by running:
   ```bash
   docker compose up -d
   
   A build process will start on the first time, you can also trigger it by running: docker compose build
   ```

### Local Endpoints

- App: http://localhost (Yes we need to use localhost to satisfy Googles Redirect URL constraints)
    - Locally the frontend runs simultaneously in Docker, so make sure it works otherwise the proxying from Backend
      to Frontend will give you headaches
- Mailpit: http://mail.sloth.localhost/
- Traefik Dashboard: http://traefik.sloth.localhost/

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

> We use Docker in Docker in production so make sure to mount your servers docker.sock into the container to test it

### Preparations

1. Make sure you have installed Docker (We recommend [Docker Engine](https://docs.docker.com/engine/)) on your server
2. Create a network called `traefik` on your server with: `docker network create traefik`
3. Run Traefik on your server via Docker (TODO: Add default traefik config and explain steps for setup)
    - A first point of orientation for now is [traefik.yml](.traefik/traefik.yml)
4. Test your build locally by running `docker build -f ./deployment/Dockerfile -t sloth/app:latest .` in the **root**
   directory
    - Make sure to change `RUN npm run generate:prod` to `RUN npm run generate` **temporarily** in
      the [release.yml](.github/workflows/release.yml) otherwise you will be redirected to the production page
5. Then you can run:
   `docker run -t -i --env-file .env -v /var/run/docker.sock:/var/run/docker.sock -p 9090:9090 --rm sloth/app:latest`

### TODO: Explain how to deploy Sloth

---

## Tests

> Make sure you are in the [root](.) directory

1. You can run all tests with `docker compose run backend go test ./...` locally
    - The [.env.test](backend/tests/.env.test) file can be used to define settings during tests

---

## Contributing 🤝

Contributions are welcome! Please submit a pull request or create an issue on the GitHub repository.

**Enjoy using Sloth! 🦥**

