package handlers

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterEndpoints(r *gin.RouterGroup) {
	// Organisation
	r.POST("organisation", h.AuthMiddleware(), h.HandleCreateOrganisation)
	r.DELETE("organisation/:id", h.AuthMiddleware(), h.HandleDeleteOrganisation)
	r.GET("organisations", h.AuthMiddleware(), h.HandleListOrganisations)
	r.GET("organisation/:id", h.AuthMiddleware(), h.HandleGetOrganisation)
	r.DELETE("organisation/member/:id/:member_id", h.AuthMiddleware(), h.HandleDeleteOrganisationMember)
	r.PUT("organisation/member/:id/:member_id", h.AuthMiddleware(), h.HandlePUTMember)
	r.PUT("organisation/member", h.AuthMiddleware(), h.HandleCreateOrganisationInvitation)
	r.POST("organisation/accept_invitation", h.AuthMiddleware(), h.HandlePOSTAcceptInvitation)
	r.DELETE("organisation/withdraw_invitation", h.AuthMiddleware(), h.HandleDELETEWithdrawInvitation)
	r.GET("organisation/:id/projects", h.AuthMiddleware(), h.HandleGETOrganisationProjects)
	r.PUT("organisation/project", h.AuthMiddleware(), h.HandlePUTOrganisationProject)
	r.DELETE("organisation/project", h.AuthMiddleware(), h.HandleRemoveProjectFromOrganisation)
	r.GET("organisation/:id/invitations", h.AuthMiddleware(), h.HandleGETInvitations)

	// Projects
	r.POST("project", h.AuthMiddleware(), h.HandleCreateProject)
	r.PUT("project/:id", h.AuthMiddleware(), h.HandleUpdateProject)
	r.GET("project/:id", h.AuthMiddleware(), h.HandleGetProject)
	r.GET("projects", h.AuthMiddleware(), h.HandleListProjects)
	r.DELETE("project/:id", h.AuthMiddleware(), h.HandleDeleteProject)
	r.GET("project/state/:id", h.AuthMiddleware(), h.HandleGetProjectState)
	r.GET("ws/project/logs/:upn/:usn", h.AuthMiddleware(), h.HandleStreamServiceLogs) // using upn and usn because depends on docker compose logs which is using the service name
	r.GET("ws/project/shell/:usn/:projectID", h.AuthMiddleware(), h.HandleStreamShell)
	// Secured by access token - don't need to chain auth-middleware
	r.GET("hook/:id", h.HandleGetProjectHook)

	// Notifications
	r.PUT("notifications", h.AuthMiddleware(), h.HandlePUTNotification)
	r.GET("notifications", h.AuthMiddleware(), h.HandleGETNotifications)

	r.PUT("user/set-current-organisation", h.AuthMiddleware(), h.SetCurrentOrganisation)

	rAuth := r.Group("auth")
	rAuth.GET(":provider", h.HandleGETAuthenticate)
	rAuth.GET(":provider/callback", h.HandleGETAuthenticateCallback)
	rAuth.GET("logout/:provider", h.HandleGETLogout)
	rAuth.GET("user", h.GetUser)
	rAuth.GET("verify-session", h.AuthMiddleware(), h.VerifyUserSession)
}
