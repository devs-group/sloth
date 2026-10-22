import type { ServiceSchema } from '~/schema/schema'

export const EmptyServiceTemplate: ServiceSchema = {
  name: '',
  ports: [''],
  image: '',
  image_tag: '',
  public: {
    enabled: false,
    hosts: [''],
    port: '',
    ssl: false,
    compress: false,
  },
  env_vars: [['', '']],
  volumes: [''],
  depends_on: {},
  deploy: {
    mode: 'replicated',
    replicas: 3,
    endpoint_mode: 'vip',
    resources: {
      limits: {
        cpus: '2.0',
        memory: '8GiB',
        pids: 100,
      },
      reservations: {
        cpus: '1.0',
        memory: '500MiB',
      },
    },
    restart_policy: {
      condition: 'on-failure',
      delay: '5s',
      max_attempts: 3,
      window: '120s',
    },
  },
  post_deploy_actions: [],
}
