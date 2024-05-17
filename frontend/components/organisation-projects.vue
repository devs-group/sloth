<script lang="ts" setup>
import type { OrganisationProject } from "~/schema/schema";

const route = useRoute();
const g = ref<OrganisationProject[]>();
 
const config = useRuntimeConfig();

</script>