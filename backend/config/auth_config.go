package config

import "os"

var AuthProviderConfig AuthProvider

type GitHubConfig struct {
	GithubClientKey       string
	GithubSecret          string
	GithubAuthCallbackURL string
}

func NewGitHubConfig() *GitHubConfig {
	return &GitHubConfig{
		GithubClientKey:       os.Getenv("GITHUB_CLIENT_KEY"),
		GithubSecret:          os.Getenv("GITHUB_SECRET"),
		GithubAuthCallbackURL: os.Getenv("GITHUB_AUTH_CALLBACK_URL"),
	}
}

type GoogleConfig struct {
	GoogleClientKey       string
	GoogleSecret          string
	GoogleAuthCallbackURL string
}

func NewGoogleConfig() *GoogleConfig {
	return &GoogleConfig{
		GoogleClientKey:       os.Getenv("GOOGLE_CLIENT_KEY"),
		GoogleSecret:          os.Getenv("GOOGLE_SECRET"),
		GoogleAuthCallbackURL: os.Getenv("GOOGLE_AUTH_CALLBACK_URL"),
	}
}

type AuthProvider struct {
	GitHubConfig *GitHubConfig
	GoogleConfig *GoogleConfig
}

func NewAuthProvider() *AuthProvider {
	return &AuthProvider{
		GitHubConfig: NewGitHubConfig(),
		GoogleConfig: NewGoogleConfig(),
	}
}
