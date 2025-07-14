import { z } from 'zod'

const RestartPolicySchema = z.object({
  condition: z.string().optional(),
  delay: z.string().optional(),
  max_attempts: z.number().optional(),
  window: z.string().optional(),
})

const ReservationsSchema = z.object({
  cpus: z.string().optional(),
  memory: z.string().optional(),
})

const LimitsSchema = z.object({
  cpus: z.string().optional(),
  memory: z.string().optional(),
  pids: z.number().optional(),
})

const ResourcesSchema = z.object({
  limits: LimitsSchema.optional(),
  reservations: ReservationsSchema.optional(),
})

const DeploySchema = z.object({
  mode: z.string().optional(),
  replicas: z.number().optional(),
  endpoint_mode: z.string().optional(),
  resources: ResourcesSchema.optional(),
  restart_policy: RestartPolicySchema.optional(),
})

const ConditionSchema = z.object({
  condition: z.string(),
})

const PostDeployActions = z.object({
  parameters: z.array(z.string()),
  shell: z.string(),
  command: z.string(),
})

export const serviceSchema = z.object({
  name: z.string().trim().min(1, 'Name is required'),
  id: z.number().optional(),
  usn: z.string().optional(),
  ports: z.array(
    z
      .string()
      .min(2, 'Minimum of 2 numbers')
      .max(6, 'Max 6 numbers')
      .regex(/^\d+$/, 'Only numbers are allowed'),
  ),
  image: z.string().trim().min(1, 'Image is required'),
  image_tag: z.string().trim().min(1, 'Image tag is required'),
  command: z.string().optional(),
  public: z.object({
    enabled: z.boolean(),
    hosts: z.array(z.string()),
    port: z.string(),
    ssl: z.boolean(),
    compress: z.boolean(),
  }),
  env_vars: z.array(
    z.array(
      z.string().refine(s => !s.includes(' '), 'Spaces are not allowed'),
    ),
  ),
  volumes: z.array(
    z.string().refine(s => !s.includes(' '), 'Spaces are not allowed'),
  ),
  healthcheck: z
    .object({
      test: z.string(),
      interval: z.string(),
      timeout: z.string(),
      retries: z.number(),
      start_period: z.string(),
    })
    .optional(),
  depends_on: z.record(ConditionSchema).optional(),
  deploy: DeploySchema.optional(),
  post_deploy_actions: z.array(PostDeployActions).optional(),
})

export const dockerCredentialSchema = z.object({
  id: z.number().optional().readonly(),
  username: z.string().trim().min(1, 'Username is required'),
  password: z.string().trim().min(1, 'Password is required'),
  registry: z.string().trim().min(1, 'Registry url is required'),
})

export const createProjectSchema = z.object({
  name: z.string().min(1, 'A project name is required ☝️🤓'),
})

export const addProjectToOrganisation = z.object({
  organisation_id: z.number().min(0),
  upn: z.string().min(1, 'A project unqiue name is required ☝️🤓'),
})

export const putMemberToOrganisation = z.object({
  organisation_id: z.number().min(0),
  email: z
    .string()
    .email('A valid E-Mail is required for the invitation  ☝️🤓'),
})

export const organisationNameSchema = z.string().min(1, 'An organisation name is required ☝️🤓')

export const projectSchema = z.object({
  id: z.number().readonly(),
  upn: z.string().optional().readonly(),
  hook: z.string().readonly(),
  access_token: z.string().readonly(),
  name: z.string().min(1, 'A project name is required ☝️🤓'),
  organisation: z.string().optional().readonly(),
  services: z.array(serviceSchema),
  docker_credentials: z.array(dockerCredentialSchema),
})

export const organisationMemberSchema = z.object({
  id: z.number().readonly(),
  userID: z.number().readonly(),
  role: z.string().readonly(),
  email: z.string(),
  username: z.string().optional(),
})

export const OrganisationMemberRoleEnum = z.enum(['owner', 'admin', 'member'], { message: 'Please select a valid role' })
export const OrganisationMemberRoleOptions = OrganisationMemberRoleEnum.options.map(role => ({
  label: role.charAt(0).toUpperCase() + role.slice(1),
  value: role,
}))

export const organisationMemberUpdateSchema = z.object({
  email: z.string().email('Invalid E-Mail 📧'),
  role: OrganisationMemberRoleEnum,
  username: z.string(),
})

export const organisationSchema = z.object({
  id: z.number().readonly(),
  organisationName: z.string().readonly(),
  currentRole: z.string().readonly(),
  members: z.array(organisationMemberSchema).optional(),
  isDefault: z.boolean().readonly(),
})

export const organisationInvitationsSchema = z.object({
  id: z.number().readonly(),
  email: z.string().readonly(),
  organisationName: z.string().readonly(),
  validUntil: z.date().readonly(),
})

export const organisationProjectSchema = z.object({
  name: z.string().readonly(),
  upn: z.string().readonly(),
  id: z.number().readonly(),
})

export const inviteToOrganisationSchema = z.string().email('A valid E-Mail is required for the invitation ☝️🤓')

export type DockerCredentialSchema = z.output<typeof dockerCredentialSchema>

export type CreateProjectType = z.output<typeof createProjectSchema>
export type ProjectSchema = z.output<typeof projectSchema>
export type Project = z.infer<typeof projectSchema>

export type ServiceSchema = z.output<typeof serviceSchema>
export type Service = z.infer<typeof serviceSchema>

export type CreateOrganisationType = z.output<typeof organisationNameSchema>
export type AddProjectToOrganisationType = z.output<
  typeof addProjectToOrganisation
>
export type InviteToOrganisationType = z.output<
  typeof inviteToOrganisationSchema
>
export type UpdateOrganisationType = z.output<typeof putMemberToOrganisation>
export type Organisation = z.infer<typeof organisationSchema>
export type OrganisationMember = z.infer<typeof organisationMemberSchema>
export type OrganisationMemberUpdate = z.infer<typeof organisationMemberUpdateSchema>
export type OrganisationProject = z.output<typeof organisationProjectSchema>
export type OrgaisationSchema = z.output<typeof organisationSchema>
export type InvitationsSchema = z.output<typeof organisationInvitationsSchema>
export type Invitation = z.infer<typeof organisationInvitationsSchema>
