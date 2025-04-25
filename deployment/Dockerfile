FROM node:20.11.0 AS frontend

WORKDIR /var/app
COPY ./frontend /var/app

RUN npm install
RUN npm run generate:prod

FROM golang:1.23.3-alpine AS backend

WORKDIR /var/app

# Trick to make use of caching
COPY go.mod go.sum /var/app/
RUN go mod download

COPY . /var/app
# Copy build files from Frontend
COPY --from=frontend /var/app/.output /var/app/frontend/.output

RUN go build -tags production -o /var/app/sloth

FROM docker:dind AS prod

WORKDIR /var/app

COPY --from=backend /var/app/database/migrations /var/app/database/migrations
COPY --from=backend /var/app/sloth /var/app/sloth

CMD ["/var/app/sloth", "run"]