package handlers

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterEndpoints(r *gin.RouterGroup) {
	// Organisation
	r.POST("organisation", AuthMiddleware(), h.HandlePOSTOrganisation)
	r.DELETE("organisation/:id", AuthMiddleware(), h.HandleDELETEOrganisation)
	r.GET("organisations", AuthMiddleware(), h.HandleGETOrganisations)

	r.GET("organisation/:id", AuthMiddleware(), h.HandleGETOrganisation)
	r.DELETE("organisation/member/:id/:member_id", AuthMiddleware(), h.HandleDELETEMember)
	r.PUT("organisation/member/:id/:member_id", AuthMiddleware(), h.HandlePUTMember)
	//r.GET("organisation/:id/:member_search", AuthMiddleware(h), h.HandleGETMembersForInvitation)

	r.GET("organisations/invitations", AuthMiddleware())
	r.PUT("organisation/member", AuthMiddleware(), h.HandlePUTInvitation)
	r.POST("organisation/accept_invitation", AuthMiddleware(), h.HandlePOSTAcceptInvitation)
	r.GET("organisation/:id/projects", AuthMiddleware(), h.HandleGetOrganisationProjects)
	r.PUT("organisation/project", AuthMiddleware(), h.HandlePUTOrganisationProject)
	r.DELETE("organisation/project", AuthMiddleware(), h.HandleDELETEOrganisationProject)
	// Projects
	r.POST("project", AuthMiddleware(), h.HandlePOSTProject)
	r.PUT("project/:upn", AuthMiddleware(), h.HandlePUTProject)
	r.GET("project/:upn", AuthMiddleware(), h.HandleGETProject)
	r.GET("projects", AuthMiddleware(), h.HandleGETProjects)
	r.DELETE("project/:upn", AuthMiddleware(), h.HandleDELETEProject)
	r.GET("project/state/:upn", AuthMiddleware(), h.HandleGETProjectState)
	r.GET("ws/project/logs/:service/:upn", AuthMiddleware(), h.HandleStreamServiceLogs)
	// Secured by access token - don't need to chain auth-middleware
	r.GET("hook/:upn", h.HandleGetProjectHook)

	rAuth := r.Group("auth")
	rAuth.GET(":provider", h.HandleGETAuthenticate)
	rAuth.GET(":provider/callback", h.HandleGETAuthenticateCallback)
	rAuth.GET("logout/:provider", h.HandleGETLogout)
	rAuth.GET("user", h.HandleGETUser)
}
