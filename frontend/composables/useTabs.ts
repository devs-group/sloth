import { shallowRef, ref, computed } from 'vue';
import DockerCredentialsForm from "~/components/docker-credentials-form.vue";
import ServicesForm from "~/components/services-form.vue";

type TabItem = {
  label: string;
  command: () => void;
  __component?: (typeof ServicesForm | typeof DockerCredentialsForm);
  disabled?: boolean;
};

export function useTabs() {
  const tabs = shallowRef<TabItem[]>([
    {
      label: "Services",
      command: () => onChangeTab(0),
      __component: ServicesForm,
    },
    {
      label: "Docker credentials",
      command: () => onChangeTab(1),
      __component: DockerCredentialsForm,
    },
    {
      label: "Monitoring (coming soon)",
      command: () => onChangeTab(2),
      disabled: true,
    },
  ]);

  const activeTabComponent = shallowRef(tabs.value[0].__component);

  function onChangeTab(idx: number) {
    if (tabs.value[idx].disabled) {
      console.log("This tab is currently disabled.");
      return;
    }
    activeTabComponent.value = tabs.value[idx].__component;
  }

  const activeTabLabel = computed(() => tabs.value.find(tab => tab.__component === activeTabComponent.value)?.label);


  return {
    tabs,
    activeTabComponent,
    onChangeTab,
    activeTabLabel
  };
}
