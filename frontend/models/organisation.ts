// Response
export interface OrganisationResponse {
  id: number
  name: string
}

// Form
export interface OrganisationPOSTFormData {
  name: string
}

export interface OrganisationPUTFormData {
  organisationName: string
}

// List Response
export type ListOrganisationResponse = OrganisationResponse[]
