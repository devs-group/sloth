import type { ServiceSchema } from '~/schema/schema'

// Single-container SeaweedFS (master, volume, filer and S3 gateway). S3 API on 8333.
// Setting the AWS_* credentials turns on S3 auth, without them the gateway is open.
export const SeaweedfsServiceTemplate: ServiceSchema = {
  name: 'seaweedfs',
  ports: ['8333'],
  image: 'chrislusf/seaweedfs',
  image_tag: '4.48',
  command: 'server -dir=/data -s3 -ip.bind=0.0.0.0',
  public: {
    enabled: false,
    hosts: [],
    port: '8333',
    ssl: false,
    compress: false,
  },
  env_vars: [
    ['AWS_ACCESS_KEY_ID', 'admin'],
    ['AWS_SECRET_ACCESS_KEY', 'change-me-please'],
  ],
  volumes: ['/data'],
  healthcheck: {
    // Filer answers and the S3 gateway (last to start, ~20s) accepts connections.
    test: 'wget -q -O /dev/null http://localhost:8888/ && nc -z localhost 8333',
    interval: '5s',
    timeout: '5s',
    retries: 5,
    start_period: '60s',
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
