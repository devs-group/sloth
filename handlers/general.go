package handlers

import (
	"crypto/rand"
	"embed"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/devs-group/sloth/config"

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
		"host":       config.Host,
	})
}

type public struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host"`
	SSL      bool   `json:"ssl"`
	Compress bool   `json:"compress"`
}

type service struct {
	Name     string     `json:"name"`
	Ports    []string   `json:"ports"`
	Image    string     `json:"image"`
	ImageTag string     `json:"image_tag"`
	Public   public     `json:"public"`
	EnvVars  [][]string `json:"env_vars"`
}

type project struct {
	Name     string    `json:"name" binding:"required"`
	Services []service `json:"services"`
}

type projectResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	UPN         string    `json:"upn"`
	AccessToken string    `json:"access_token"`
	Hook        string    `json:"hook"`
	Services    []service `json:"services"`
}

func (h *Handler) HandlePOSTProject(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req project
	if err := c.BindJSON(&req); err != nil {
		slog.Error("unable to parse request body", "err", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	accessTokenLen := 12
	accessToken, err := randStringRunes(accessTokenLen)
	if err != nil {
		slog.Error("unable to generate access token", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	uniqueProjectSuffixLen := 10
	upnSuffix, err := randStringRunes(uniqueProjectSuffixLen)
	if err != nil {
		slog.Error("unable to generate unique project name suffix", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	upn := fmt.Sprintf("%s-%s", req.Name, upnSuffix)

	dc := generateDockerCompose(req, upn)

	dcj, err := dc.ToJSONString()
	if err != nil {
		slog.Error("unable to parse docker compose struct to json string", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	projectsDir := config.ProjectsDir
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

	projectDir := path.Join(config.ProjectsDir, upn)
	err = h.store.InsertProjectWithTx(u.UserID, req.Name, upn, accessToken, dcj, projectDir, func() error {
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

func (h *Handler) HandlePUTProject(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var req project
	if err := c.BindJSON(&req); err != nil {
		slog.Error("unable to parse request body", "err", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	upn := c.Param("upn")

	dc := generateDockerCompose(req, upn)

	dcj, err := dc.ToJSONString()
	if err != nil {
		slog.Error("unable to parse docker compose to json string", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	dcy, err := dc.ToYAML()
	if err != nil {
		slog.Error("unable to parse docker compose to yaml string", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ppath := getProjectPath(upn)

	p, err := h.store.UpdateProjectWithTx(u.UserID, upn, req.Name, dcj, func() error {
		err := renameDockerComposeFile(upn)
		if err != nil {
			slog.Error("unable to rename docker-compose file to temp", "upn", upn, "err", err)
			return err
		}

		err = createDockerComposeFile(upn, dcy)
		if err != nil {
			slog.Error("unable to rename create a new docker-compose file", "upn", upn, "err", err)
			return err
		}

		_, err = startContainers(ppath)
		if err != nil {
			slog.Error("unable to restart containers", "upn", upn, "err", err)
			return err
		}

		err = deleteDockerComposeTempFile(upn)
		if err != nil {
			slog.Error("unable to delete docker-compose file", "upn", upn, "err", err)
			return err
		}
		return nil
	})

	if err != nil {
		err := deleteDockerComposeFile(upn)
		if err != nil {
			slog.Error("unable to delete docker-compose file", "upn", upn, "err", err)
		}

		err = rollbackRenameDockerComposeFile(upn)
		if err != nil {
			slog.Error("unable to rollback rename of docker-compose file", "upn", upn, "err", err)
		}

		_, err = startContainers(ppath)
		if err != nil {
			slog.Error("unable to restart containers", "upn", upn, "err", err)
		}

		slog.Error("unable to update project", "upn", upn, "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	pr, err := createProjectResponse(p)
	if err != nil {
		slog.Error("unable to create project response struct", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, pr)
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

	pp := getProjectPath(p.UniqueName)
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

	r := make([]*projectResponse, 0, len(projects))
	for i := range projects {
		p := projects[i]
		pr, err := createProjectResponse(&p)
		if err != nil {
			slog.Error("unable to create project response struct", "err", err)
			continue
		}
		if pr != nil {
			r = append(r, pr)
		} else {
			slog.Error("something went wrong while creating the project response struct")
			continue
		}
	}
	c.JSON(http.StatusOK, r)
}

func (h *Handler) HandleGETProject(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	upn := c.Param("upn")
	p, err := h.store.SelectProjectByUPN(u.UserID, upn)
	if err != nil {
		slog.Error("unable to select project", "upn", upn, "err", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	pr, err := createProjectResponse(p)
	if err != nil {
		slog.Error("unable to create project response struct", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, pr)
}

func createProjectResponse(p *database.Project) (*projectResponse, error) {
	dc, err := compose.FromString(p.DCJ)
	if err != nil {
		slog.Error("unable to parse docker compose json string", "err", err)
		return nil, err
	}

	ppath := getProjectPath(p.UniqueName)
	_, err = getContainersState(ppath)
	if err != nil {
		slog.Error("unable to get containers status", "err", err)
		return nil, err
	}

	services := make([]service, 0)

	for k, s := range dc.Services {
		host, err := s.Labels.GetHost()
		if err != nil {
			slog.Error("unable to get host from labels", "err", err)
		}
		image := strings.Split(s.Image, ":")
		if len(image) < 2 {
			return nil, fmt.Errorf("unsuported image, expected 'image:tag' format got: %s", s.Image)
		}
		var envVars [][]string
		for _, e := range s.Environment {
			kv := strings.Split(e, "=")
			envVars = append(envVars, kv)
		}
		services = append(services, service{
			Name:     k,
			Ports:    s.Ports,
			Image:    image[0],
			ImageTag: image[1],
			EnvVars:  envVars,
			Public: public{
				Enabled:  s.Labels.IsPublic(),
				Host:     host,
				SSL:      s.Labels.IsSSL(),
				Compress: s.Labels.IsCompress(),
			},
		})
	}
	return &projectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Services:    services,
		UPN:         p.UniqueName,
		AccessToken: p.AccessToken,
		Hook:        fmt.Sprintf("%s/v1/hook/%s", config.Host, p.UniqueName),
	}, nil
}

func randStringRunes(n int) (string, error) {
	var runes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(runes))))
		if err != nil {
			return "", err
		}
		b[i] = runes[n.Int64()]
	}
	return string(b), nil
}

func createDockerComposeFile(upn, yaml string) error {
	p := fmt.Sprintf("%s/%s/%s", filepath.Clean(config.ProjectsDir), upn, "docker-compose.yml")
	filePerm := 0600
	err := os.WriteFile(p, []byte(yaml), os.FileMode(filePerm))
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
			Ports:    s.Ports,
		}

		if len(s.EnvVars) > 0 {
			for _, v := range s.EnvVars {
				c.Environment = append(c.Environment, fmt.Sprintf("%s=%s", v[0], v[1]))
			}
		}

		if s.Public.Enabled {
			usn := fmt.Sprintf("%s-%s", upn, s.Name)
			host := strings.ToLower(fmt.Sprintf("%s.devs-group.ch", usn))

			if s.Public.Host != "" {
				host = strings.ToLower(s.Public.Host)
			}

			hasHostEnv := false
			for _, e := range c.Environment {
				if strings.HasPrefix(e, "HOST=") {
					hasHostEnv = true
				}
			}

			if !hasHostEnv {
				c.Environment = append(c.Environment, fmt.Sprintf("HOST=%s", host))
			}

			labels := []string{
				"traefik.enable=true",
				fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%s", usn, s.Ports[0]),
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
	if config.Environment == config.Development {
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
	for i := range containers {
		c := containers[i]
		sn := c.Labels["com.docker.compose.service"]
		state[sn] = containerState{
			State:  c.State,
			Status: c.Status,
		}
	}
	return state, nil
}

func getProjectPath(upn string) string {
	return path.Join(filepath.Clean(config.ProjectsDir), upn)
}

func renameDockerComposeFile(upn string) error {
	oldPath := path.Join(filepath.Clean(config.ProjectsDir), upn, "docker-compose.yml")
	newPath := path.Join(filepath.Clean(config.ProjectsDir), upn, "docker-compose.yml.tmp")
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}

// rollbackRenameDockerComposeFile renames docker-compose.yml.tmp file to docker-compose.yml file
func rollbackRenameDockerComposeFile(upn string) error {
	tmpPath := path.Join(filepath.Clean(config.ProjectsDir), upn, "docker-compose.yml.tmp")
	newPath := path.Join(filepath.Clean(config.ProjectsDir), upn, "docker-compose.yml")
	_, err := os.Stat(tmpPath)
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, newPath)
}

func deleteDockerComposeFile(upn string) error {
	tmpPath := path.Join(filepath.Clean(config.ProjectsDir), upn, "docker-compose.yml")
	_, err := os.Stat(tmpPath)
	if err != nil {
		return err
	}
	return os.Remove(tmpPath)
}

func deleteDockerComposeTempFile(upn string) error {
	tmpPath := path.Join(filepath.Clean(config.ProjectsDir), upn, "docker-compose.yml.tmp")
	_, err := os.Stat(tmpPath)
	if err != nil {
		return err
	}
	return os.Remove(tmpPath)
}
