import type { ProjectSchema } from '~/schema/schema';
import { PreDefinedServices } from '~/schema/schema';
export function useService(p: Ref<ProjectSchema | undefined>) {
  function addService(predefinedServiceKey: String | null ) {
    const service = PreDefinedServices.get(predefinedServiceKey ?? "")
    if ( service ) {
        p.value?.services.push(JSON.parse(JSON.stringify(service)));
    }
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
      registry: ""
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
  };
}
