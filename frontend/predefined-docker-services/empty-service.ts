import type { ServiceSchema } from "~/schema/schema";
export const EmptyService: ServiceSchema = {
    name: "",
    ports: [""],
    image: "",
    image_tag: "",
    public: {
      enabled: false,
      hosts: [""],
      port: "",
      ssl: true,
      compress: false,
    },
    env_vars: [["", ""]],
    volumes: [""],
    healthcheck: {
      test: ["CMD-SHELL", "curl -f http://localhost/ || exit 1"],
      interval: "30s",
      timeout: "10s",
      retries: 3,
      start_period: "15s",
    },
    depends_on: {},
    deploy: {
      mode: "replicated",
      replicas: 3,
      endpoint_mode: "vip",
      resources: {
        limits: {
          cpus: "2.0",
          memory: "8GiB",
          pids: 100
        },
        reservations: {
          cpus: "1.0",
          memory: "500MiB"
        }
      },
      restart_policy: {
        condition: "on-failure",
        delay: "5s",
        max_attempts: 3,
        window: "120s"
      },
    }
  };