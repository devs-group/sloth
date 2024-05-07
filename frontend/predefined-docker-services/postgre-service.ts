import type { ServiceSchema } from "~/schema/schema";

export const DockerPostgreService: ServiceSchema = {
    name: "postgres-service",
    ports: ["5432"],
    image: "postgres",
    image_tag: "13",
    public: {
      enabled: false,
      hosts: [],
      port: "5432",
      ssl: false,
      compress: true
    },
    env_vars: [
      ["POSTGRES_DB", "exampledb"],
      ["POSTGRES_USER", "exampleuser"],
      ["POSTGRES_PASSWORD", "examplepass"],
    ],
    volumes: [],
    healthcheck: {
      test: ["CMD-SHELL", "pg_isready -U exampleuser"],
      interval: "30s",
      timeout: "10s",
      retries: 5,
      start_period: "15s"
    },
    deploy: {
      resources: {
        limits: {
          cpus: "0.5",
          memory: "512M"
        },
        reservations: {
          cpus: "0.25",
          memory: "256M"
        }
      },
      restart_policy: {
        condition: "on-failure",
        delay: "5s",
        max_attempts: 3,
        window: "120s"
      }
    }
  };