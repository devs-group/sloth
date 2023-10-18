import {z} from "zod";

export const serviceSchema = z.object({
    name: z.string(),
    ports: z.array(
        z.string().min(2, "Minimum of 2 numbers").max(6, "Max 6 numbers").regex(/^\d+$/, "Only numbers are allowed")
    ),
    image: z.string(),
    image_tag: z.string(),
    public: z.object({
        enabled: z.boolean(),
        host: z.string(),
        ssl: z.boolean(),
        compress: z.boolean()
    }),
    env_vars: z.array(
        z.tuple([
            z.string().refine(s => !s.includes(' '), 'Spaces are not allowed'),
            z.string().refine(s => !s.includes(' '), 'Spaces are not allowed')
        ]))
})

export const projectSchema = z.object({
    id: z.number().optional().readonly(),
    upn: z.string().optional().readonly(),
    hook: z.string().optional().readonly(),
    access_token: z.string().optional().readonly(),
    name: z.string().refine(s => !s.includes(' '), 'Spaces are not allowed'),
    services: z.array(serviceSchema)
})

export type ProjectSchema = z.output<typeof projectSchema>
export type ServiceSchema = z.output<typeof serviceSchema>
export type Project = z.infer<typeof projectSchema>
export type Service = z.infer<typeof serviceSchema>
