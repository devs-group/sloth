import { z } from "zod";

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

export const groupSchema = z.object({
  group_name: z.string().readonly(),
  members: z.array(z.string()).optional(),
});

export type ProjectSchema = z.output<typeof projectSchema>;
export type ServiceSchema = z.output<typeof serviceSchema>;
export type DockerCredentialSchema = z.output<typeof dockerCredentialSchema>;
export type GroupSchema = z.output<typeof groupSchema>;

export type Group = z.infer<typeof groupSchema>;
export type Project = z.infer<typeof projectSchema>;
export type Service = z.infer<typeof serviceSchema>;
