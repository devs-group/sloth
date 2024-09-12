import ServiceShellDialog from '~/components/dialogs/service-shell-dialog.vue';
import { PreDefinedServices } from '~/schema/schema';
import {DialogProps} from "~/config/const";
import { EmptyServiceTemplate } from "~/service-templates/empty-service-template";
import { useWebSocket } from "@vueuse/core";
import type { IServiceState } from "~/config/interfaces";
import type { Project, ServiceSchema } from "~/schema/schema";

import type { IServiceState } from '~/config/interfaces';
import type { Project } from '~/schema/schema';

export function useService(p: Ref<Project | null>) {
  const config = useRuntimeConfig();
  const isShellModalOpen = ref(false)

  function addService(service: ServiceSchema) {
    if (service && p.value) {
      p.value.services = [...p.value.services, structuredClone(service)];
    }
    return service;
  }

  function addEnv(serviceIdx: number) {
    p.value?.services[serviceIdx].env_vars.push(["", ""]);
  }

  function removeEnv(envIdx: number, serviceIdx: number) {
    p.value?.services[serviceIdx].env_vars.splice(envIdx, 1);
  }

  function addVolume(serviceIdx: number) {
    p.value?.services[serviceIdx].volumes.push("");
  }

  function removeVolume(volIdx: number, serviceIdx: number) {
    p.value?.services[serviceIdx].volumes.splice(volIdx, 1);
  }

  function addPort(serviceIdx: number) {
    p.value?.services[serviceIdx].ports.push("");
  }

  function removePort(portIdx: number, serviceIdx: number) {
    p.value?.services[serviceIdx].ports.splice(portIdx, 1);
  }

  function removeService(idx: number) {
    p.value?.services.splice(idx, 1);
  }

  function addCredential() {
    p.value?.docker_credentials.push({
      username: "",
      password: "",
      registry: "",
    });
  }

  function removeCredential(idx: number) {
    p.value?.docker_credentials.splice(idx, 1);
  }

  function addHost(serviceIdx: number) {
    p.value?.services[serviceIdx].public.hosts.push("");
  }

  function removeHost(hostIdx: number, serviceIdx: number) {
    p.value?.services[serviceIdx].public.hosts.splice(hostIdx, 1);
  }

  function addPostDeployAction(serviceIdx: number) {
    p.value?.services[serviceIdx].post_deploy_actions?.push({
      parameters: [],
      shell: "",
      command: "",
    });
  }

  function removePostDeployAction(
    postDeployActionIdx: number,
    serviceIdx: number,
  ) {
    p.value?.services[serviceIdx].post_deploy_actions?.splice(
      postDeployActionIdx,
      1,
    );
  }

  async function fetchServiceStates(id: string) {
    return $fetch<Record<string, IServiceState>>(
      `${config.public.backendHost}/v1/project/state/${id}`,
      {
        method: "GET",
        credentials: "include",
      },
    );
  }

  function streamServiceLogs(upn: string, usn: string, logsLines: String[]) {
    const wsBackendHost = config.public.backendHost.replace("http", "ws");
    const { status, data, close } = useWebSocket(
      `${wsBackendHost}/v1/ws/project/logs/${upn}/${usn}`,
      {
        autoReconnect: {
          retries: 5,
          delay: 1000,
          onFailed() {
            console.log("ERROR");
          },
        },
      },
    );

    watchEffect(() => {
      logsLines?.push(data.value);
    });
  }

  function startServiceShell(id: number, serviceName: string, dialog: any) {

    const wsBackendHost = config.public.backendHost.replace("http", "ws")
    const {status, close, send, data, open} = useWebSocket(
      `${wsBackendHost}/v1/ws/project/shell/${serviceName}/${id}`,
      {
        autoReconnect: {
          retries: 5,
          delay: 1000,
          onFailed() {
              console.log("ERROR")
          },
        },
        immediate: false
      }
    );

    open()
    dialog.open(ServiceShellDialog, {
      props: {
        header: 'Terminal',
        ...DialogProps.BigDialog,
        closable: true
      },
      data: {
        send: send,
        data: data
      },
      onClose() {
        isShellModalOpen.value = false
        close()
      }
    })
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
    fetchServiceStates,
    streamServiceLogs,
    startServiceShell,
  };
}
