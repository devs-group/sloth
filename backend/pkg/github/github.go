package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type GitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

type GitHubSearchResults struct {
	TotalCount        int          `json:"total_count"`
	IncompleteResults bool         `json:"incomplete_results"`
	Items             []GitHubUser `json:"items"`
}

// TODO: Talk if this isn't better to get called on the frontend
func SearchGitHubUsers(query string) (*GitHubSearchResults, error) {
	url := fmt.Sprintf("https://api.github.com/search/users?q=%s", query)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error making request:", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			slog.Error("Error closing body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("GitHub API returned non-200 status:", resp.Status)
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil, err
	}

	var searchResults GitHubSearchResults
	if err := json.Unmarshal(body, &searchResults); err != nil {
		fmt.Println("Error parsing JSON:", err)
		return nil, err
	}

	return &searchResults, nil
}
