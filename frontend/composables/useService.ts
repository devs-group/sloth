import { useWebSocket, type UseWebSocketReturn } from '@vueuse/core'
import type { Project, ServiceSchema } from '~/schema/schema'

export function useService(p: Ref<Project | null>) {
  const config = useRuntimeConfig()

  function addService(service: ServiceSchema) {
    if (service && p.value) {
      service.public.ssl = true
      p.value.services = [...p.value.services, structuredClone(service)]
    }
    return service
  }

  function addEnv(serviceIdx: number) {
    p.value?.services[serviceIdx].env_vars.push(['', ''])
  }

  function removeEnv(envIdx: number, serviceIdx: number) {
    p.value?.services[serviceIdx].env_vars.splice(envIdx, 1)
  }

  function addVolume(serviceIdx: number) {
    p.value?.services[serviceIdx].volumes.push('')
  }

  function removeVolume(volIdx: number, serviceIdx: number) {
    p.value?.services[serviceIdx].volumes.splice(volIdx, 1)
  }

  function addPort(serviceIdx: number) {
    p.value?.services[serviceIdx].ports.push('')
  }

  function removePort(portIdx: number, serviceIdx: number) {
    p.value?.services[serviceIdx].ports.splice(portIdx, 1)
  }

  function removeService(idx: number) {
    p.value?.services.splice(idx, 1)
  }

  function addCredential() {
    p.value?.docker_credentials.push({
      username: '',
      password: '',
      registry: '',
    })
  }

  function removeCredential(idx: number) {
    p.value?.docker_credentials.splice(idx, 1)
  }

  function addHost(serviceIdx: number) {
    p.value?.services[serviceIdx].public.hosts.push('')
  }

  function removeHost(hostIdx: number, serviceIdx: number) {
    p.value?.services[serviceIdx].public.hosts.splice(hostIdx, 1)
  }

  function addPostDeployAction(serviceIdx: number) {
    p.value?.services[serviceIdx].post_deploy_actions?.push({
      parameters: [],
      shell: '',
      command: '',
    })
  }

  function removePostDeployAction(
    postDeployActionIdx: number,
    serviceIdx: number,
  ) {
    p.value?.services[serviceIdx].post_deploy_actions?.splice(
      postDeployActionIdx,
      1,
    )
  }

  function streamServiceLogs(
    upn: string,
    usn: string,
  ): UseWebSocketReturn<string> {
    const wsBackendHost = config.public.backendHost.replace('http', 'ws')
    return useWebSocket(`${wsBackendHost}/v1/ws/project/logs/${upn}/${usn}`, {
      autoReconnect: {
        retries: 5,
        delay: 1000,
        onFailed() {
          console.log('ERROR')
        },
      },
    })
  }

  function startServiceShell(
    projectID: number,
    usn: string,
  ): UseWebSocketReturn<string> {
    const wsBackendHost = config.public.backendHost.replace('http', 'ws')
    return useWebSocket(
      `${wsBackendHost}/v1/ws/project/shell/${usn}/${projectID}`,
      {
        autoReconnect: {
          retries: 5,
          delay: 1000,
          onFailed() {
            console.log('ERROR')
          },
        },
      },
    )
  }

  return {
    addService,
    addEnv,
    removeEnv,
    addVolume,
    removeVolume,
    addPort,
    removePort,
    removeService,
    addCredential,
    removeCredential,
    addHost,
    removeHost,
    addPostDeployAction,
    removePostDeployAction,
    streamServiceLogs,
    startServiceShell,
  }
}
