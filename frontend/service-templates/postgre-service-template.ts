import type { ServiceSchema } from '~/schema/schema'

export const PostgreServiceTemplate: ServiceSchema = {
  name: 'PostgreSQL',
  ports: ['5432'],
  image: 'postgres',
  image_tag: '13',
  public: [
    {
      enabled: false,
      host: '',
      port: '5432',
      ssl: true,
      compress: false,
    },
  ],
  env_vars: [
    ['POSTGRES_DB', 'exampledb'],
    ['POSTGRES_USER', 'exampleuser'],
    ['POSTGRES_PASSWORD', 'examplepass'],
  ],
  volumes: [],
  healthcheck: {
    test: 'pg_isready -U exampleuser',
    interval: '30s',
    timeout: '10s',
    retries: 5,
    start_period: '15s',
  },
  deploy: {
    resources: {
      limits: {
        cpus: '0.5',
        memory: '512M',
      },
      reservations: {
        cpus: '0.25',
        memory: '256M',
      },
    },
    restart_policy: {
      condition: 'on-failure',
      delay: '5s',
      max_attempts: 3,
      window: '120s',
    },
  },
}
