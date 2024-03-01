package handlers

import "github.com/gin-gonic/gin"

func (h *Handler) RegisterEndpoints(r *gin.RouterGroup) {
	// Group
	r.POST("group", AuthMiddleware(), h.HandlePOSTGroup)
	r.DELETE("group/:group_name", AuthMiddleware(), h.HandleDELETEGroup)
	r.GET("groups", AuthMiddleware(), h.HandleGETGroups)

	r.GET("group/:group_name", AuthMiddleware(), h.HandleGETGroup)
	r.DELETE("group/member/:group_name/:member_id", AuthMiddleware(), h.HandleDELETEMember)
	r.PUT("group/member/:group_name/:member_id", AuthMiddleware(), h.HandlePUTMember)
	//r.GET("group/:group_name/:member_search", AuthMiddleware(), h.HandleGETMembersForInvitation)

	r.GET("groups/invitations", AuthMiddleware())
	r.PUT("group/member", AuthMiddleware(), h.HandlePUTInvitation)
	r.POST("group/accept_invitation", AuthMiddleware(), h.HandlePOSTAcceptInvitation)
	r.GET("group/:group_name/projects", AuthMiddleware(), h.HandleGetGroupProjects)
	r.PUT("group/project", AuthMiddleware(), h.HandlePUTGroupProject)
	r.DELETE("group/project", AuthMiddleware(), h.HandleDELETEGroupProject)
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
