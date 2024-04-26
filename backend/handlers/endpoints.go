package handlers

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterEndpoints(r *gin.RouterGroup) {
	// Organization
	r.POST("organization", AuthMiddleware(), h.HandlePOSTOrganization)
	r.DELETE("organization/:organization_name", AuthMiddleware(), h.HandleDELETEOrganization)
	r.GET("organizations", AuthMiddleware(), h.HandleGETOrganizations)

	r.GET("organization/:organization_name", AuthMiddleware(), h.HandleGETOrganization)
	r.DELETE("organization/member/:organization_name/:member_id", AuthMiddleware(), h.HandleDELETEMember)
	r.PUT("organization/member/:organization_name/:member_id", AuthMiddleware(), h.HandlePUTMember)
	//r.GET("organization/:organization_name/:member_search", AuthMiddleware(), h.HandleGETMembersForInvitation)

	r.GET("organizations/invitations", AuthMiddleware())
	r.PUT("organization/member", AuthMiddleware(), h.HandlePUTInvitation)
	r.POST("organization/accept_invitation", AuthMiddleware(), h.HandlePOSTAcceptInvitation)
	r.GET("organization/:organization_name/projects", AuthMiddleware(), h.HandleGetOrganizationProjects)
	r.PUT("organization/project", AuthMiddleware(), h.HandlePUTOrganizationProject)
	r.DELETE("organization/project", AuthMiddleware(), h.HandleDELETEOrganizationProject)
	// Projects
	r.POST("project", AuthMiddleware(), h.HandlePOSTProject)
	r.PUT("project/:upn", AuthMiddleware(), h.HandlePUTProject)
	r.GET("project/:upn", AuthMiddleware(), h.HandleGETProject)
	r.GET("projects", AuthMiddleware(), h.HandleGETProjects)
	r.DELETE("project/:upn", AuthMiddleware(), h.HandleDELETEProject)
	r.GET("project/state/:upn", AuthMiddleware(), h.HandleGETProjectState)
	r.GET("ws/project/logs/:service/:upn", AuthMiddleware(), h.HandleStreamServiceLogs)
	// Secured by access token - dont need to chain auth-middleware
	r.GET("hook/:upn", h.HandleGetProjectHook)

	rAuth := r.Group("auth")
	rAuth.GET(":provider", h.HandleGETAuthenticate)
	rAuth.GET(":provider/callback", h.HandleGETAuthenticateCallback)
	rAuth.GET("logout/:provider", h.HandleGETLogout)
	rAuth.GET("user", h.HandleGETUser)
}
