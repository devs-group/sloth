package handlers

import (
	"embed"
	"fmt"
	"github.com/devs-group/sloth/config"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/devs-group/sloth/database"
	"github.com/devs-group/sloth/pkg/compose"
	"github.com/devs-group/sloth/pkg/compose/docker"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	store    *database.Store
	vueFiles embed.FS
}

func NewHandler(store *database.Store, vueFiles embed.FS) Handler {
	return Handler{
		store:    store,
		vueFiles: vueFiles,
	}
}

func (h *Handler) HandleGETInfo(c *gin.Context) {
	var version string
	err := database.DB.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)
	if err != nil {
		slog.Error("unable to query sqlite version", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"sqlite_ver": version,
		"host":       config.HOST,
	})
}

type public struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	SSL      bool   `json:"ssl"`
	Compress bool   `json:"compress"`
}

type service struct {
	Name     string `json:"name"`
	Port     string `json:"port"`
	Image    string `json:"image"`
	ImageTag string `json:"image_tag"`
	Public   public
	EnvVars  map[string]string `json:"env_vars"`
}

type project struct {
	Name     string    `json:"name" binding:"required"`
	Services []service `json:"services"`
}

func (h *Handler) HandlePOSTProject(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var p project
	if err := c.BindJSON(&p); err != nil {
		slog.Error("unable to parse request body", "err", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	accessToken := randStringRunes(12)
	upn := fmt.Sprintf("%s-%s", p.Name, randStringRunes(6))

	dc := generateDockerCompose(p, upn)

	dcj, err := dc.ToJSONString()
	if err != nil {
		slog.Error("unable to parse docker compose struct to json string", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	projectsDir := config.PROJECTS_DIR
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		if err := os.MkdirAll(projectsDir, os.ModePerm); err != nil {
			slog.Error("failed to create folder", "path", projectsDir, "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		} else {
			slog.Debug("folder created successfully", "path", projectsDir)
		}
	} else if err != nil {
		slog.Error("unable to check if folder exists", "path", projectsDir, "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else {
		slog.Debug("folder already exists", "path", projectsDir)
	}

	projectDir := path.Join(config.PROJECTS_DIR, upn)
	err = h.store.InsertProjectWithTx(u.UserID, p.Name, upn, accessToken, dcj, projectDir, func() error {
		err = os.Mkdir(projectDir, os.ModePerm)
		if err != nil {
			slog.Error("unable to create directory", "dir", projectDir, "err", err)
			return err
		}

		yaml, err := dc.ToYAML()
		if err != nil {
			slog.Error("unable to to parse docker-compose to yaml", "err", err)
			return err
		}

		err = createDockerComposeFile(upn, yaml)
		if err != nil {
			slog.Error("unable to create docker-compose.yml file", "err", err)
			return err
		}
		slog.Info("created project", "dir", projectDir)
		return nil
	})
	if err != nil {
		slog.Error("unable to create project", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":              "ok",
		"access_token":        accessToken,
		"unique_project_name": upn,
	})
}

func (h *Handler) HandleGETHook(ctx *gin.Context) {
	upn := ctx.Param("unique_project_name")
	accessToken := ctx.GetHeader("X-Access-Token")
	if accessToken == "" {
		slog.Error("X-Access-Token header is required")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	p, err := h.store.GetProjectByNameAndAccessToken(upn, accessToken)
	if err != nil {
		slog.Error("unable to find project by name and access token", "err", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	slog.Info("executing restart script...")

	pp := getProjectPath(p)
	containers, err := startContainers(pp)
	if err != nil {
		slog.Error("unable to execute startup script", "err", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"upn":        p.UniqueName,
		"containers": containers,
	})
}

func (h *Handler) HandleGETProjects(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	projects, err := h.store.SelectProjects(u.UserID)
	if err != nil {
		slog.Error("unable to select projects", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	type res struct {
		ID          int              `json:"id"`
		Name        string           `json:"name"`
		UPN         string           `json:"upn"`
		AccessToken string           `json:"access_token"`
		Hook        string           `json:"hook"`
		Services    map[string]gin.H `json:"services"`
	}
	r := make([]res, 0, len(projects))
	for _, p := range projects {
		dc, err := compose.FromString(p.DCJ)
		if err != nil {
			slog.Error("unable to parse docker compose json string", "err", err)
			continue
		}

		ppath := getProjectPath(&p)
		containers, err := getContainersState(ppath)
		if err != nil {
			slog.Error("unable to get containers status", "err", err)
			continue
		}

		services := make(map[string]gin.H)
		for k, s := range dc.Services {
			services[k] = gin.H{
				"name":     k,
				"ports":    s.Ports,
				"image":    s.Image,
				"env_vars": s.Environment,
				"status":   containers[k].Status,
				"state":    containers[k].State,
			}
		}
		r = append(r, res{
			ID:          p.ID,
			Name:        p.Name,
			Services:    services,
			UPN:         p.UniqueName,
			AccessToken: p.AccessToken,
			Hook:        fmt.Sprintf("%s/v1/hook/%s", config.HOST, p.UniqueName),
		})
	}
	c.JSON(http.StatusOK, r)
}

func randStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func createDockerComposeFile(upn string, yaml string) error {
	p := fmt.Sprintf("%s/%s/%s", filepath.Clean(config.PROJECTS_DIR), upn, "docker-compose.yml")
	err := os.WriteFile(p, []byte(yaml), 0777)
	if err != nil {
		return fmt.Errorf("unable to write file %s: err %v", p, err)
	}
	return nil
}

func generateDockerCompose(p project, upn string) compose.DockerCompose {
	services := make(map[string]*compose.Container)
	for _, s := range p.Services {
		c := &compose.Container{
			Image:    fmt.Sprintf("%s:%s", s.Image, s.ImageTag),
			Restart:  "always",
			Networks: []string{"web", "default"},
			Ports:    []string{s.Port},
		}

		if len(s.EnvVars) > 0 {
			for k, v := range s.EnvVars {
				c.Environment = append(c.Environment, fmt.Sprintf("%s=%s", k, v))
			}
		}

		if s.Public.Enabled {
			usn := fmt.Sprintf("%s-%s", upn, s.Name)
			host := strings.ToLower(fmt.Sprintf("%s.devs-group.ch", usn))

			if s.Public.Host != "" {
				host = strings.ToLower(s.Public.Host)
			}

			c.Environment = append(c.Environment, fmt.Sprintf("HOST=%s", host))

			labels := []string{
				"traefik.enable=true",
				fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%s", usn, s.Port),
				fmt.Sprintf("traefik.http.routers.%s.rule=Host(`%s`)", usn, host),
			}

			if s.Public.SSL {
				labels = append(
					labels,
					fmt.Sprintf("traefik.http.routers.%s.entrypoints=https", usn),
					fmt.Sprintf("traefik.http.routers.%s.tls=true", usn),
					fmt.Sprintf("traefik.http.routers.%s.tls.certresolver=le", usn),
				)
			}

			if s.Public.Compress {
				labels = append(
					labels,
					fmt.Sprintf("traefik.http.middlewares.%s-compress.compress=true", usn),
					fmt.Sprintf("traefik.http.routers.%s.middlewares=%s-compress", usn, usn),
				)
			}

			c.Labels = labels
		}

		services[s.Name] = c
	}

	// External networks refer to pre-existing networks on the host machine.
	// In a production environment, this network is typically established during Traefik setup.
	// However, in development environments, this network may not be present by default.
	isWebExternalNetwork := true
	if config.ENVIRONMENT == config.Development {
		isWebExternalNetwork = false
	}

	dc := compose.DockerCompose{
		Version: "3.9",
		Networks: map[string]*compose.Network{
			"web": {
				External: isWebExternalNetwork,
			},
			"default": {
				Driver:   "bridge",
				External: false,
			},
		},
		Services: services,
	}

	return dc
}

type containerState struct {
	State  string `json:"state"`
	Status string `json:"status"`
}

func startContainers(ppath string) (map[string]containerState, error) {
	if err := compose.Pull(ppath); err != nil {
		return nil, err
	}
	if err := compose.Down(ppath); err != nil {
		return nil, err
	}
	if err := compose.Up(ppath); err != nil {
		return nil, err
	}
	return getContainersState(ppath)
}

func getContainersState(ppath string) (map[string]containerState, error) {
	containers, err := docker.GetContainersByDirectory(ppath)
	if err != nil {
		return nil, err
	}
	state := make(map[string]containerState)
	for _, c := range containers {
		sn := c.Labels["com.docker.compose.service"]
		state[sn] = containerState{
			State:  c.State,
			Status: c.Status,
		}
	}
	return state, nil
}

func getProjectPath(p *database.Project) string {
	return fmt.Sprintf("%s/%s", filepath.Clean(config.PROJECTS_DIR), p.UniqueName)
}
