import { z } from "zod";
import { EmptyService } from "~/predefined-docker-services/empty-service";
import { DockerPostgreService } from "~/predefined-docker-services/postgre-service";

const RestartPolicySchema = z.object({
  condition: z.string().optional(),
  delay: z.string().optional(),
  max_attempts: z.number().optional(),
  window: z.string().optional(),
});

const ReservationsSchema = z.object({
  cpus: z.string().optional(),
  memory: z.string().optional(),
});

const LimitsSchema = z.object({
  cpus: z.string().optional(),
  memory: z.string().optional(),
  pids: z.number().optional(),
});

const ResourcesSchema = z.object({
  limits: LimitsSchema.optional(),
  reservations: ReservationsSchema.optional(),
});

const DeploySchema = z.object({
  mode: z.string().optional(),
  replicas: z.number().optional(),
  endpoint_mode: z.string().optional(),
  resources: ResourcesSchema.optional(),
  restart_policy: RestartPolicySchema.optional(),
});

const ConditionSchema = z.object({
  condition: z.string(),
});

export const serviceSchema = z.object({
  name: z.string(),
  usn: z.string().optional(),
  ports: z.array(
    z
      .string()
      .min(2, "Minimum of 2 numbers")
      .max(6, "Max 6 numbers")
      .regex(/^\d+$/, "Only numbers are allowed")
  ),
  image: z.string(),
  image_tag: z.string(),
  command: z.string().optional(),
  public: z.object({
    enabled: z.boolean(),
    hosts: z.array(z.string()),
    port: z.string(),
    ssl: z.boolean(),
    compress: z.boolean(),
  }),
  env_vars: z.array(
    z.tuple([
      z.string().refine((s) => !s.includes(" "), "Spaces are not allowed"),
      z.string().refine((s) => !s.includes(" "), "Spaces are not allowed"),
    ])
  ),
  volumes: z.array(
    z.string().refine((s) => !s.includes(" "), "Spaces are not allowed")
  ),
  healthcheck: z.object({
    test: z.array(z.string()),
    interval: z.string(),
    timeout: z.string(),
    retries: z.number(),
    start_period: z.string(),
  }),
  depends_on: z.record(ConditionSchema).optional(),
  deploy: DeploySchema.optional(),
});

export const dockerCredentialSchema = z.object({
  id: z.number().optional().readonly(),
  username: z.string().trim().min(1, "Username is required"),
  password: z.string().trim().min(1, "Password is required"),
  registry: z.string().trim().min(1, "Registry url is required"),
});

export const projectSchema = z.object({
  id: z.number().optional().readonly(),
  upn: z.string().optional().readonly(),
  hook: z.string().optional().readonly(),
  access_token: z.string().optional().readonly(),
  name: z.string(),
  group: z.string().optional().readonly(),
  services: z.array(serviceSchema),
  docker_credentials: z.array(dockerCredentialSchema),
});

export const organizationSchema = z.object({
  organization_name: z.string().readonly(),
  is_owner: z.boolean().optional(),
  members: z.array(z.string()).optional(),
});

export const invitationsSchema = z.object({
  organization_name: z.string().readonly(),
  user_id: z.string().readonly(),
});

export const GroupProject = z.object({
  name: z.string().readonly(),
  upn: z.string().readonly(),
});

export type ProjectSchema = z.output<typeof projectSchema>;
export type ServiceSchema = z.output<typeof serviceSchema>;
export type DockerCredentialSchema = z.output<typeof dockerCredentialSchema>;
export type GroupProject = z.output<typeof GroupProject>;
export type GroupSchema = z.output<typeof organizationSchema>;
export type InvitationsSchema = z.output<typeof invitationsSchema>;

export type Invitation = z.infer<typeof invitationsSchema>;
export type Group = z.infer<typeof organizationSchema>;
export type Project = z.infer<typeof projectSchema>;
export type Service = z.infer<typeof serviceSchema>;

export const PreDefinedServices: Map<String,ServiceSchema> = new Map([
  ["", EmptyService],
  ["Postgres", DockerPostgreService]
]);
