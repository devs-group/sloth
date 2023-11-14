package handlers

import (
	"crypto/rand"
	"embed"
	"fmt"
	"log"
	"log/slog"
	"math/big"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/devs-group/sloth/pkg/docker"
	"github.com/goombaio/namegenerator"
	"github.com/gorilla/websocket"

	"github.com/devs-group/sloth/config"

	"github.com/devs-group/sloth/database"
	"github.com/devs-group/sloth/pkg/compose"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	store    *database.Store
	vueFiles embed.FS
	upgrader websocket.Upgrader
}

func New(store *database.Store, vueFiles embed.FS) Handler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	// TODO: Loop over list of trusted origins instead returning true for all origins.
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return Handler{
		store:    store,
		vueFiles: vueFiles,
		upgrader: upgrader,
	}
}

type public struct {
	Enabled  bool   `json:"enabled"`
	Host     string `json:"host" binding:"required"`
	Port     string `json:"port" binding:"required,numeric"`
	SSL      bool   `json:"ssl"`
	Compress bool   `json:"compress"`
}

type service struct {
	Name     string     `json:"name" binding:"required"`
	Ports    []string   `json:"ports" binding:"gt=0"`
	Image    string     `json:"image" binding:"required"`
	ImageTag string     `json:"image_tag" binding:"required"`
	Command  string     `json:"command"`
	Public   public     `json:"public"`
	EnvVars  [][]string `json:"env_vars"`
	Volumes  []string   `json:"volumes" binding:"dive,dirpath"`
}

type project struct {
	ID                int                `json:"id"`
	UPN               string             `json:"upn"`
	AccessToken       string             `json:"access_token"`
	Hook              string             `json:"hook"`
	Name              string             `json:"name" binding:"required"`
	Services          []service          `json:"services"`
	DockerCredentials []dockerCredential `json:"docker_credentials"`
}

type dockerCredential struct {
	ID       int    `json:"id"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Registry string `json:"registry" binding:"required,uri"`
}

const persistentVolumeDirectoryName = "data"
const dockerComposeFileName = "docker-compose.yml"
const dockerConfigFileName = "config.json"

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
	upn := fmt.Sprintf("%s-%s", generateRandomName(), upnSuffix)

	projectsDir := config.ProjectsDir
	_, err = createFolderIfNotExists(projectsDir)
	if err != nil {
		slog.Error("unable to create folder", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ppath := getProjectPath(upn)

	var volumesPath string
	if hasVolumesInRequest(req) {
		volumesPath, err = createFolderIfNotExists(path.Join(ppath, persistentVolumeDirectoryName))
		if err != nil {
			slog.Error("unable to create folder", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	dc := generateDockerCompose(req, upn, volumesPath)

	dcj, err := dc.ToJSONString()
	if err != nil {
		slog.Error("unable to parse docker compose struct to json string", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var dockerCreds []database.DockerCredential
	for _, dc := range req.DockerCredentials {
		dockerCreds = append(dockerCreds, database.DockerCredential{
			Username: dc.Username,
			Password: dc.Password,
			Registry: dc.Registry,
		})
	}

	err = h.store.InsertProjectWithTx(u.UserID, req.Name, upn, accessToken, dcj, ppath, dockerCreds, func() error {
		_, err = createFolderIfNotExists(ppath)
		if err != nil {
			slog.Error("unable to create directory", "dir", ppath, "err", err)
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

		slog.Info("created project", "dir", ppath)
		return nil
	})
	if err != nil {
		slog.Error("unable to create project", "err", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
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

	ppath := getProjectPath(upn)

	var volumesPath string
	if hasVolumesInRequest(req) {
		volumesPath, err = createFolderIfNotExists(path.Join(ppath, persistentVolumeDirectoryName))
		if err != nil {
			slog.Error("unable to create folder", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	dc := generateDockerCompose(req, upn, volumesPath)

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

	var dockerCreds []database.DockerCredential
	for _, dc := range req.DockerCredentials {
		dockerCreds = append(dockerCreds, database.DockerCredential{
			ID:       dc.ID,
			Username: dc.Username,
			Password: dc.Password,
			Registry: dc.Registry,
		})
	}

	p, err := h.store.UpdateProjectWithTx(u.UserID, upn, req.Name, dcj, dockerCreds, func() error {
		// Create temp files
		err := createTempFile(dockerComposeFileName, upn)
		if err != nil {
			slog.Error("unable to rename docker-compose file to temp", "upn", upn, "err", err)
			return err
		}

		err = createTempFile(dockerConfigFileName, upn)
		if err != nil {
			slog.Error("unable to rename config file to temp", "upn", upn, "err", err)
			return err
		}

		// Create configuration files
		err = createDockerComposeFile(upn, dcy)
		if err != nil {
			slog.Error("unable to rename create a new docker-compose file", "upn", upn, "err", err)
			return err
		}

		// Restart project
		err = restartContainers(ppath, dc.Services, req.DockerCredentials)
		if err != nil {
			slog.Error("unable to restart containers", "upn", upn, "err", err)
			return err
		}

		// Delete temp files
		err = deleteFile(fmt.Sprintf("%s.tmp", dockerComposeFileName), upn)
		if err != nil {
			slog.Error("unable to delete docker-compose file", "upn", upn, "err", err)
			return err
		}

		err = deleteFile(fmt.Sprintf("%s.tmp", dockerConfigFileName), upn)
		if err != nil {
			slog.Error("unable to delete docker-compose file", "upn", upn, "err", err)
			return err
		}
		return nil
	})

	if err != nil {
		slog.Error("unable to update project", "upn", upn, "err", err)
		// Rollback
		err := deleteFile(dockerComposeFileName, upn)
		if err != nil {
			slog.Error("unable to delete docker-compose file", "upn", upn, "err", err)
		}

		err = rollbackFromTempFile(dockerComposeFileName, upn)
		if err != nil {
			slog.Error("unable to rollback rename of docker-compose file", "upn", upn, "err", err)
		}

		err = deleteFile(dockerConfigFileName, upn)
		if err != nil {
			slog.Error("unable to delete docker config file", "upn", upn, "err", err)
		}

		err = rollbackFromTempFile(dockerConfigFileName, upn)
		if err != nil {
			slog.Error("unable to rollback rename of docker config file", "upn", upn, "err", err)
		}

		// Restart project
		err = restartContainers(ppath, dc.Services, req.DockerCredentials)
		if err != nil {
			slog.Error("unable to restart containers", "upn", upn, "err", err)
		}

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
	upn := ctx.Param("upn")
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

	pp := getProjectPath(p.UniqueName)

	dc, err := compose.FromString(p.DCJ)
	if err != nil {
		slog.Error("unable to find parse docker compose from string", "err", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	dcs := make([]dockerCredential, len(p.DockerCredentials))
	for i, dc := range p.DockerCredentials {
		dcs[i] = dockerCredential{Username: dc.Username, Password: dc.Password, Registry: dc.Registry}
	}

	err = restartContainers(pp, dc.Services, dcs)
	if err != nil {
		slog.Error("unable to execute startup script", "err", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"upn": p.UniqueName,
	})
}

func (h *Handler) HandleGETProjectState(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	upn := c.Param("upn")

	p, err := h.store.SelectProjectByUPN(u.UserID, upn)
	if err != nil || p == nil {
		slog.Error("unable to find project by upn", "upn", upn, "err", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	state, err := getContainersState(upn)
	if err != nil {
		slog.Error("unable to get project state", "upn", upn, "err", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, state)
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

	r := make([]*project, len(projects))
	for i := range projects {
		p := projects[i]
		pr, err := createProjectResponse(&p)
		if err != nil {
			slog.Error("unable to create project response struct", "err", err)
			continue
		}
		if pr != nil {
			r[i] = pr
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

func (h *Handler) HandleDELETEProject(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	upn := c.Param("upn")
	ppath := getProjectPath(upn)
	deletedProjectPath := fmt.Sprintf("%s-deleted", ppath)
	err = h.store.DeleteProjectByUPNWithTx(u.UserID, upn, func() error {
		return renameFolder(ppath, deletedProjectPath)
	})
	if err != nil {
		slog.Error("unable to delete project. trying to roll back...", "upn", upn, "err", err)
		err = renameFolder(deletedProjectPath, ppath)
		if err != nil {
			slog.Error("unable to rename folder", "err", err)
		}
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Delete the temp folder in background
	go func() {
		err := deleteFolder(deletedProjectPath)
		if err != nil {
			slog.Error("unable to delete folder", "path", deletedProjectPath, "err", err)
		}
	}()

	c.Status(http.StatusOK)
}

func (h *Handler) HandleStreamServiceLogs(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	upn := c.Param("upn")
	s := c.Param("service")

	p, err := h.store.SelectProjectByUPN(u.UserID, upn)
	if err != nil || p == nil {
		slog.Error("unable to find project by upn", "upn", upn, "err", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("unable to upgrade http to ws")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("unable to close websocket connection", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}(conn)

	ppath := getProjectPath(upn)
	out := make(chan string)
	go func() {
		err := compose.Logs(ppath, s, out)
		if err != nil {
			slog.Error("unable to stream logs", "upn", upn, "service", s, "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}()

	line := 0
	for o := range out {
		line++
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%d %s", line, o)))
	}
}

func (h *Handler) HandleStreamServiceShell(c *gin.Context) {
	u, err := getUserFromSession(c.Request)
	if err != nil {
		slog.Error("unable to get user from session", "err", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	upn := c.Param("upn")
	s := c.Param("service")

	p, err := h.store.SelectProjectByUPN(u.UserID, upn)
	if err != nil || p == nil {
		slog.Error("unable to find project by upn", "upn", upn, "err", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("unable to upgrade http to ws")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			slog.Error("unable to close websocket connection", "err", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}(conn)

	out := make(chan []byte)
	in := make(chan []byte)

	go compose.Exec(upn, s, in, out)

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			in <- message
		}
	}()

	for o := range out {
		_ = conn.WriteMessage(websocket.TextMessage, o)
	}
}

func createProjectResponse(p *database.Project) (*project, error) {
	dc, err := compose.FromString(p.DCJ)
	if err != nil {
		slog.Error("unable to parse docker compose json string", "err", err)
		return nil, err
	}

	services := make([]service, len(dc.Services))
	idx := 0
	for k, s := range dc.Services {
		host, err := s.Labels.GetHost()
		if err != nil {
			slog.Error("unable to get host from labels", "err", err)
		}
		image := strings.Split(s.Image, ":")
		if len(image) < 2 {
			return nil, fmt.Errorf("unsuported image, expected 'image:tag' format got: %s", s.Image)
		}

		envVars := make([][]string, len(s.Environment))
		for i, e := range s.Environment {
			kv := strings.Split(e, "=")
			envVars[i] = kv
		}

		// When no env vars are set, response with empty tuple
		if len(s.Environment) == 0 {
			envVars = [][]string{{"", ""}}
		}

		volumes := make([]string, len(s.Volumes))
		for i, v := range s.Volumes {
			volumes[i] = strings.Split(v, ":")[1]
		}

		// When no volumes are set, response with empty string
		if len(s.Volumes) == 0 {
			volumes = []string{""}
		}

		port, err := s.Labels.GetPort()
		if err != nil {
			slog.Error("unable to get port from labels", "err", err)
		}

		services[idx] = service{
			Name:     k,
			Ports:    s.Ports,
			Command:  s.Command,
			Image:    image[0],
			ImageTag: image[1],
			EnvVars:  envVars,
			Volumes:  volumes,
			Public: public{
				Enabled:  s.Labels.IsPublic(),
				Host:     host,
				Port:     port,
				SSL:      s.Labels.IsSSL(),
				Compress: s.Labels.IsCompress(),
			},
		}
		idx++
	}

	dockerCreds := make([]dockerCredential, len(p.DockerCredentials))
	for i, dc := range p.DockerCredentials {
		dockerCreds[i] = dockerCredential{
			ID:       dc.ID,
			Username: dc.Username,
			Password: dc.Password,
			Registry: dc.Registry,
		}
	}

	return &project{
		ID:                p.ID,
		Name:              p.Name,
		Services:          services,
		UPN:               p.UniqueName,
		AccessToken:       p.AccessToken,
		Hook:              fmt.Sprintf("%s/v1/hook/%s", config.Host, p.UniqueName),
		DockerCredentials: dockerCreds,
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
	p := fmt.Sprintf("%s/%s/%s", filepath.Clean(config.ProjectsDir), upn, dockerComposeFileName)
	filePerm := 0600
	err := os.WriteFile(p, []byte(yaml), os.FileMode(filePerm))
	if err != nil {
		return fmt.Errorf("unable to write file %s: err %v", p, err)
	}
	return nil
}

func generateDockerCompose(p project, upn string, volumesPath string) compose.DockerCompose {
	services := make(map[string]*compose.Container)
	for _, s := range p.Services {
		c := &compose.Container{
			Image:    fmt.Sprintf("%s:%s", s.Image, s.ImageTag),
			Restart:  "always",
			Networks: []string{"web", "default"},
			Ports:    s.Ports,
		}

		if s.Command != "" {
			c.Command = s.Command
		}

		for _, ev := range s.EnvVars {
			if len(ev) == 2 && ev[0] != "" && ev[1] != "" {
				c.Environment = append(c.Environment, fmt.Sprintf("%s=%s", ev[0], ev[1]))
			}
		}

		if len(s.Volumes) > 0 && volumesPath != "" && s.Volumes[0] != "" {
			for _, v := range s.Volumes {
				c.Volumes = append(c.Volumes, fmt.Sprintf("./%s:%s", persistentVolumeDirectoryName, v))
			}
		}

		if s.Public.Enabled {
			usn := fmt.Sprintf("%s-%s", upn, s.Name)
			host := strings.ToLower(fmt.Sprintf("%s.devs-group.ch", usn))

			if s.Public.Host != "" {
				host = strings.ToLower(s.Public.Host)
			}

			labels := []string{
				"traefik.enable=true",
				fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%s", usn, s.Public.Port),
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

func restartContainers(ppath string, services compose.Services, credentials []dockerCredential) error {
	err := runDockerLogin(ppath, credentials)
	if err != nil {
		slog.Error("unable to run docker login", "path", ppath, "err", err)
		return err
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(services))
	for _, s := range services {
		wg.Add(1)
		go func(service *compose.Container) {
			defer wg.Done()
			err := docker.Pull(service.Image, ppath)
			if err != nil {
				errCh <- err
			}
		}(s)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	var errors []error
	for err := range errCh {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return fmt.Errorf("unable to pull containers: %v", errors)
	}

	for _, dc := range credentials {
		err = docker.Logout(ppath, dc.Registry)
		if err != nil {
			return fmt.Errorf("unable to run docker logout for registry%s: %v", dc.Registry, err)
		}
	}

	if err := compose.Down(ppath); err != nil {
		return fmt.Errorf("unable to shut down containers: %v", err)
	}

	if err := compose.Up(ppath); err != nil {
		return fmt.Errorf("unable to start containers: %v", err)
	}

	return nil
}

func getContainersState(upn string) (map[string]containerState, error) {
	containers, err := docker.GetContainersByDirectory(getProjectPath(upn))
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

func createTempFile(filename, upn string) error {
	oldPath := path.Join(filepath.Clean(config.ProjectsDir), upn, filename)
	newPath := path.Join(filepath.Clean(config.ProjectsDir), upn, fmt.Sprintf("%s.tmp", filename))
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	err := os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}
	return nil
}

// rollbackFromTempFile renames filename.tmp file to filename file
func rollbackFromTempFile(filename, upn string) error {
	tmpPath := path.Join(filepath.Clean(config.ProjectsDir), upn, fmt.Sprintf("%s.tmp", filename))
	newPath := path.Join(filepath.Clean(config.ProjectsDir), upn, filename)
	_, err := os.Stat(tmpPath)
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, newPath)
}

func deleteFile(filename, upn string) error {
	p := path.Join(filepath.Clean(config.ProjectsDir), upn, filename)
	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return os.Remove(p)
}

func createFolderIfNotExists(p string) (string, error) {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		if err := os.MkdirAll(p, os.ModePerm); err != nil {
			return "", fmt.Errorf("failed to crete folder in path %s, err: %v", p, err)
		} else {
			slog.Debug("folder has been created successfully", "path", p)
			return p, nil
		}
	} else if err != nil {
		return "", fmt.Errorf("unable to check if folder exists in path %s, err: %v", p, err)
	} else {
		slog.Debug("folder already exists", "path", p)
		return p, nil
	}
}

func hasVolumesInRequest(p project) bool {
	hasVolumes := false
	for _, s := range p.Services {
		if len(s.Volumes) > 0 {
			hasVolumes = true
		}
	}
	return hasVolumes
}

func renameFolder(oldPath, newPath string) error {
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	return nil
}

func deleteFolder(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

func generateRandomName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)
	return nameGenerator.Generate()
}

func runDockerLogin(ppath string, credentials []dockerCredential) error {
	if len(credentials) == 0 {
		return nil
	}
	for _, dc := range credentials {
		err := docker.Login(dc.Username, dc.Password, dc.Registry, ppath)
		if err != nil {
			return err
		}
	}
	return nil
}
