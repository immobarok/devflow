package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mobarok/devflow/pkg/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type UserProfile struct {
	Provider   string
	ProviderID string
	Email      string
	FirstName  string
	LastName   string
	AvatarURL  string
}

type Provider interface {
	GetAuthURL(state string) string
	ExchangeAndFetchProfile(ctx context.Context, code string) (*UserProfile, error)
}

type Manager struct {
	providers map[string]Provider
}

func NewManager(cfg *config.Config) *Manager {
	providers := make(map[string]Provider)

	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		providers["google"] = &googleProvider{
			config: &oauth2.Config{
				ClientID:     cfg.GoogleClientID,
				ClientSecret: cfg.GoogleClientSecret,
				RedirectURL:  cfg.GoogleRedirectURL,
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
					"https://www.googleapis.com/auth/userinfo.profile",
				},
				Endpoint: google.Endpoint,
			},
		}
	}

	if cfg.GithubClientID != "" && cfg.GithubClientSecret != "" {
		providers["github"] = &githubProvider{
			config: &oauth2.Config{
				ClientID:     cfg.GithubClientID,
				ClientSecret: cfg.GithubClientSecret,
				RedirectURL:  cfg.GithubRedirectURL,
				Scopes: []string{
					"read:user",
					"user:email",
				},
				Endpoint: github.Endpoint,
			},
		}
	}

	return &Manager{providers: providers}
}

func (m *Manager) GetProvider(name string) (Provider, error) {
	p, ok := m.providers[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("oauth provider '%s' is not configured", name)
	}
	return p, nil
}

// Google Provider implementation
type googleProvider struct {
	config *oauth2.Config
}

func (g *googleProvider) GetAuthURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *googleProvider) ExchangeAndFetchProfile(ctx context.Context, code string) (*UserProfile, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange google code: %w", err)
	}

	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		ID         string `json:"id"`
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
		FamilyName string `json:"family_name"`
		Picture    string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode google userinfo: %w", err)
	}

	if data.Email == "" {
		return nil, fmt.Errorf("google userinfo did not return an email")
	}

	return &UserProfile{
		Provider:   "google",
		ProviderID: data.ID,
		Email:      data.Email,
		FirstName:  data.GivenName,
		LastName:   data.FamilyName,
		AvatarURL:  data.Picture,
	}, nil
}

// GitHub Provider implementation
type githubProvider struct {
	config *oauth2.Config
}

func (gh *githubProvider) GetAuthURL(state string) string {
	return gh.config.AuthCodeURL(state)
}

func (gh *githubProvider) ExchangeAndFetchProfile(ctx context.Context, code string) (*UserProfile, error) {
	token, err := gh.config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange github code: %w", err)
	}

	client := gh.config.Client(ctx, token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch github user: %w", err)
	}
	defer resp.Body.Close()

	var data struct {
		ID        int64  `json:"id"`
		Name      string `json:"name"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode github user: %w", err)
	}

	email := data.Email
	// If public email is empty, fetch private emails from /user/emails
	if email == "" {
		email = gh.fetchPrimaryEmail(client)
	}

	if email == "" {
		return nil, fmt.Errorf("github account did not return a verified email")
	}

	firstName := data.Name
	lastName := ""
	parts := strings.SplitN(data.Name, " ", 2)
	if len(parts) == 2 {
		firstName = parts[0]
		lastName = parts[1]
	}
	if firstName == "" {
		firstName = data.Login
	}

	return &UserProfile{
		Provider:   "github",
		ProviderID: strconv.FormatInt(data.ID, 10),
		Email:      email,
		FirstName:  firstName,
		LastName:   lastName,
		AvatarURL:  data.AvatarURL,
	}, nil
}

func (gh *githubProvider) fetchPrimaryEmail(client *http.Client) string {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}

	if err := json.Unmarshal(body, &emails); err != nil {
		return ""
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email
		}
	}
	if len(emails) > 0 {
		return emails[0].Email
	}
	return ""
}
