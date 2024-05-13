import { z } from "zod";
import { EmptyServiceTemplate } from "~/service-templates/empty-service-template";
import { PostgreServiceTemplate } from "~/service-templates/postgre-service-template";

const UserSchema = z.object({
  avatar_url: z.string().optional(),
  email: z.string().optional(),
  first_name: z.string().optional(),
  id: z.number().readonly(),
  last_name: z.string().optional(),
  location: z.string().optional(),
  name: z.string().optional(),
  nickname: z.string().optional(), 
});

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

export const organisationSchema = z.object({
  organisation_name: z.string().readonly(),
  organisation_id: z.number().optional(),
  is_owner: z.boolean().optional(),
  members: z.array(z.string()).optional(),
});

export const organisationInvitationsSchema = z.object({
  organisation_name: z.string().readonly(),
  user_id: z.string().readonly(),
});

export const GroupProject = z.object({
  name: z.string().readonly(),
  upn: z.string().readonly(),
});

export type DockerCredentialSchema = z.output<typeof dockerCredentialSchema>;

export type ProjectSchema = z.output<typeof projectSchema>;
export type Project = z.infer<typeof projectSchema>;

export type ServiceSchema = z.output<typeof serviceSchema>;
export type Service = z.infer<typeof serviceSchema>;

export type UserSchema = z.infer<typeof UserSchema>;
export type Organisation = z.infer<typeof organisationSchema>;
export type OrganisationProject = z.output<typeof GroupProject>;
export type OrgaisationSchema = z.output<typeof organisationSchema>;
export type InvitationsSchema = z.output<typeof organisationInvitationsSchema>;
export type Invitation = z.infer<typeof organisationInvitationsSchema>;

export const PreDefinedServices: Map<String,ServiceSchema> = new Map([
  ["", EmptyServiceTemplate],
  ["Postgres", PostgreServiceTemplate]
]);
