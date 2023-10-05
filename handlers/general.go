package handlers

import (
	"deployer/database"
	"deployer/pkg/compose"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

var projectsDir = getEnv("PROJECTS_DIR", "/Users/robert/Projects/deployer/test_projects")

const restartScript = `
#!/bin/sh

echo "Pulling new containers";
docker-compose pull;

echo "Shutting down containers";
docker-compose down;

echo "Starting containers";
docker-compose up -d;
`
const restartScriptName = "restart.sh"

func HandleGETInfo(c *gin.Context) {
	var version string
	err := database.DB.QueryRow("SELECT SQLITE_VERSION()").Scan(&version)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"sqlite_ver": version,
	})
}

type public struct {
	Enabled  bool   `json:"enabled"`
	URL      string `json:"url"`
	SSL      bool   `json:"ssl"`
	Compress bool   `json:"compress"`
}

type service struct {
	Name     string `json:"name"`
	Port     string `json:"port"`
	Image    string `json:"image"`
	ImageTag string `json:"image_tag"`
	Public   public
}

type project struct {
	Name     string    `json:"name" binding:"required"`
	Services []service `json:"services"`
}

func HandlePOSTProject(ctx *gin.Context) {
	var p project
	if err := ctx.BindJSON(&p); err != nil {
		slog.Error("unable to parse request body", "err", err)
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	accessToken := randStringRunes(12)
	upn := fmt.Sprintf("%s-%s", p.Name, randStringRunes(6))

	// beginning db transaction and committing only when files on the file system have been created successfully
	tx, err := database.DB.Begin()
	defer tx.Rollback()
	if err != nil {
		slog.Error("unable to begin sql transaction")
		return
	}

	dc := generateDockerCompose(p, upn)

	dcj, err := dc.ToJSONString()
	if err != nil {
		slog.Error("unable to parse docker compose struct to json string", "err", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	_, err = tx.Exec(
		"INSERT INTO projects (unique_name, dcj, access_token) VALUES ($1, json($2), $3)", upn, dcj, accessToken)
	if err != nil {
		slog.Error("unable to execute query", "err", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	projectDir := path.Join(projectsDir, upn)
	out, err := exec.Command("mkdir", projectDir).Output()
	if err != nil {
		slog.Error("unable to create directory", "dir", projectDir)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = createRestartScript(upn)
	if err != nil {
		slog.Error("unable to create restart script", "err", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	yaml, err := dc.ToYAML()
	if err != nil {
		slog.Error("unable to to parse docker-compose to yaml", "err", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = createDockerComposeFile(upn, yaml)
	if err != nil {
		slog.Error("unable to create docker-compose.yml file", "err", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	tx.Commit()

	slog.Info("created project", "dir", projectDir, "out", string(out))

	ctx.JSON(http.StatusOK, gin.H{
		"status":              "ok",
		"access_token":        accessToken,
		"unique_project_name": upn,
	})
}

func HandleGETHook(ctx *gin.Context) {
	upn := ctx.Param("unique_project_name")
	accessToken := ctx.GetHeader("X-Access-Token")
	if accessToken == "" {
		ctx.AbortWithError(http.StatusUnauthorized, fmt.Errorf("X-Access-Token header is required"))
		return
	}

	type project struct {
		ID         int    `db:"id"`
		UniqueName string `db:"unique_name"`
		DCJ        string `db:"dcj"`
	}
	var p project
	err := database.DB.Get(&p, "SELECT id, unique_name, dcj FROM projects WHERE unique_name = $1 AND access_token = $2", upn, accessToken)
	if err != nil {
		slog.Error("unable to execute query", "err", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	path := fmt.Sprintf("%s/%s/%s", filepath.Clean(projectsDir), upn, restartScript)
	cmd, err := exec.Command("/bin/sh", path).Output()
	if err != nil {
		slog.Error("unable to execute command", "cmd", "/bin/sh "+path, "err", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	slog.Info("exucuted command", "stdout", string(cmd))

	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"upn":    upn,
	})
}

func randStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func createRestartScript(upn string) error {
	path := fmt.Sprintf("%s/%s/%s", filepath.Clean(projectsDir), upn, restartScriptName)
	err := os.WriteFile(path, []byte(restartScript), 0777)
	if err != nil {
		return fmt.Errorf("unable to write file %s: err %v", path, err)
	}
	return nil
}

func createDockerComposeFile(upn string, yaml string) error {
	path := fmt.Sprintf("%s/%s/%s", filepath.Clean(projectsDir), upn, "docker-compose.yml")
	err := os.WriteFile(path, []byte(yaml), 0777)
	if err != nil {
		return fmt.Errorf("unable to write file %s: err %v", path, err)
	}
	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func generateDockerCompose(p project, upn string) compose.DockerCompose {
	services := make(map[string]*compose.Container)
	for _, s := range p.Services {
		c := &compose.Container{
			Image:    fmt.Sprintf("%s:%s", s.Image, s.ImageTag),
			Restart:  "always",
			Networks: []string{"web", "default"},
		}

		if s.Public.Enabled {
			usn := fmt.Sprintf("%s-%s", upn, s.Name)
			url := strings.ToLower(fmt.Sprintf("%s.devs-group.ch", usn))

			if s.Public.URL != "" {
				url = strings.ToLower(s.Public.URL)
			}

			labels := []string{
				"traefik.enable=true",
				fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port=%s", usn, s.Port),
				fmt.Sprintf("traefik.http.routers.%s.rule=Host(`%s`)", usn, url),
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

	dc := compose.DockerCompose{
		Version: "3.9",
		Networks: map[string]*compose.Network{
			"web": {
				External: true,
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
