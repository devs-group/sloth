import { initCustomFormatter } from "vue";
import { z } from "zod";

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
    test: z.string(),
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

export const createProjectSchema = z.object({
  name: z.string().min(1, "A project name is required ☝️🤓"),
});

export const projectSchema = z.object({
  id: z.number().readonly(),
  upn: z.string().optional().readonly(),
  hook: z.string().readonly(),
  access_token: z.string().readonly(),
  name: z.string().min(1, "A project name is required ☝️🤓"),
  organisation: z.string().optional().readonly(),
  services: z.array(serviceSchema),
  docker_credentials: z.array(dockerCredentialSchema),
});

export const organisationSchema = z.object({
  id: z.string().readonly(),
  organisation_name: z.string().readonly(),
  is_owner: z.boolean().optional(),
  members: z.array(z.string()).optional(),
});

export const invitationsSchema = z.object({
  organisation_name: z.string().readonly(),
  user_id: z.string().readonly(),
});

export const organisationProjectSchema = z.object({
  name: z.string().readonly(),
  upn: z.string().readonly(),
});

export type CreateProjectSchema = z.output<typeof createProjectSchema>;
export type ProjectSchema = z.output<typeof projectSchema>;
export type ServiceSchema = z.output<typeof serviceSchema>;
export type DockerCredentialSchema = z.output<typeof dockerCredentialSchema>;
export type OrganisationProjectSchema = z.output<typeof organisationProjectSchema>;
export type OrganisationSchema = z.output<typeof organisationSchema>;
export type InvitationsSchema = z.output<typeof invitationsSchema>;

export type Invitation = z.infer<typeof invitationsSchema>;
export type Organisation = z.infer<typeof organisationSchema>;
export type Project = z.infer<typeof projectSchema>;
export type Service = z.infer<typeof serviceSchema>;
