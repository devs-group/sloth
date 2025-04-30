import type { ServiceSchema } from '~/schema/schema'

export const MinioServiceTemplate: ServiceSchema = {
  name: 'MinIO',
  ports: ['9000', '9001'],
  image: 'minio/minio',
  image_tag: 'latest',
  command: 'server /data --console-address ":9001"',
  public: [
    {
      enabled: true,
      host: '',
      port: '9001',
      ssl: true,
      compress: false,
    },
  ],
  env_vars: [
    ['MINIO_ROOT_USER', 'admin'],
    ['MINIO_ROOT_PASSWORD', 'admin123'],
  ],
  volumes: ['/data'],
  healthcheck: {
    test: 'curl -f http://localhost:9000/minio/health/live || exit 1',
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
