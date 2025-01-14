FROM node:20.6.1 AS frontend

WORKDIR /app
COPY ./frontend /app

RUN npm install
RUN npm run generate:prod

FROM golang:1.22.2-alpine AS backend

WORKDIR /app

# Trick to make use of caching
COPY go.mod go.sum /app/
RUN go mod download

COPY . /app
# Copy build files from Frontend
COPY --from=frontend /app/.output /app/frontend/.output

RUN go build -tags production -o /app/sloth

FROM docker:dind AS prod

WORKDIR /app

COPY --from=backend /app/database/migrations /app/database/migrations
COPY --from=backend /app/sloth /app/sloth

CMD ["/app/sloth", "run"]