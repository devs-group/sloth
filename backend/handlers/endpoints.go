package handlers

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterEndpoints(r *gin.RouterGroup) {
	// Organisation
	r.POST("organisation", h.AuthMiddleware(), h.HandlePOSTOrganisation)
	r.DELETE("organisation/:id", h.AuthMiddleware(), h.HandleDELETEOrganisation)
	r.GET("organisations", h.AuthMiddleware(), h.HandleGETOrganisations)

	r.GET("organisation/:id", h.AuthMiddleware(), h.HandleGETOrganisation)
	r.DELETE("organisation/member/:id/:member_id", h.AuthMiddleware(), h.HandleDELETEMember)
	r.PUT("organisation/member/:id/:member_id", h.AuthMiddleware(), h.HandlePUTMember)
	//r.GET("organisation/:id/:member_search", AuthMiddleware(h), h.HandleGETMembersForInvitation)

	r.GET("organisations/invitations", h.AuthMiddleware())
	r.PUT("organisation/member", h.AuthMiddleware(), h.HandlePUTInvitation)
	r.POST("organisation/accept_invitation", h.AuthMiddleware(), h.HandlePOSTAcceptInvitation)
	r.GET("organisation/:id/projects", h.AuthMiddleware(), h.HandleGETOrganisationProjects)
	r.PUT("organisation/project", h.AuthMiddleware(), h.HandlePUTOrganisationProject)
	r.DELETE("organisation/project", h.AuthMiddleware(), h.HandleDELETEOrganisationProject)
	// Projects
	r.POST("project", h.AuthMiddleware(), h.HandlePOSTProject)
	r.PUT("project/:id", h.AuthMiddleware(), h.HandlePUTProject)
	r.GET("project/:id", h.AuthMiddleware(), h.HandleGETProject)
	r.GET("projects", h.AuthMiddleware(), h.HandleGETProjects)
	r.DELETE("project/:id", h.AuthMiddleware(), h.HandleDELETEProject)
	r.GET("project/state/:id", h.AuthMiddleware(), h.HandleGETProjectState)
	r.GET("ws/project/logs/:service/:id", h.AuthMiddleware(), h.HandleStreamServiceLogs)
	r.GET("ws/project/shell/:service/:id", h.AuthMiddleware(), h.HandleStreamShell)
	// Secured by access token - don't need to chain auth-middleware
	r.GET("hook/:id", h.HandleGetProjectHook)

	rAuth := r.Group("auth")
	rAuth.GET(":provider", h.HandleGETAuthenticate)
	rAuth.GET(":provider/callback", h.HandleGETAuthenticateCallback)
	rAuth.GET("logout/:provider", h.HandleGETLogout)
	rAuth.GET("user", h.HandleGETUser)
	rAuth.GET("verify-session", h.AuthMiddleware(), h.HandleGETVerifySession)
}
