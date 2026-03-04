package auth

import (
	"context"
	"fmt"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

// OIDCProvider wraps OpenID Connect authentication.
type OIDCProvider struct {
	provider *oidc.Provider
	verifier *oidc.IDTokenVerifier
	config   *oauth2.Config
}

// NewOIDCProvider creates a new OIDC provider.
func NewOIDCProvider(ctx context.Context, config *OAuth2Config) (*OIDCProvider, error) {
	// Discover OIDC provider
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("create OIDC provider: %w", err)
	}

	// Create ID token verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.ClientID,
	})

	// OAuth2 config
	oauth2Cfg := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		RedirectURL:  config.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       config.Scopes,
	}

	if len(oauth2Cfg.Scopes) == 0 {
		oauth2Cfg.Scopes = []string{oidc.ScopeOpenID, "profile", "email"}
	}

	return &OIDCProvider{
		provider: provider,
		verifier: verifier,
		config:   oauth2Cfg,
	}, nil
}

// GetAuthURL generates the OIDC authorization URL.
func (p *OIDCProvider) GetAuthURL(state string) string {
	return p.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges authorization code for tokens.
func (p *OIDCProvider) ExchangeCode(ctx context.Context, code string) (*oauth2.Token, *oidc.IDToken, error) {
	oauth2Token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return nil, nil, fmt.Errorf("exchange code: %w", err)
	}

	// Extract ID token
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return nil, nil, fmt.Errorf("no id_token in response")
	}

	// Verify ID token
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, nil, fmt.Errorf("verify id_token: %w", err)
	}

	return oauth2Token, idToken, nil
}

// ClaimsWithEntitlements holds standard OIDC claims plus groups/roles.
type ClaimsWithEntitlements struct {
	Email         string   `json:"email"`
	EmailVerified bool     `json:"email_verified"`
	Name          string   `json:"name"`
	Picture       string   `json:"picture"`
	Groups        []string `json:"groups"`
	Roles         []string `json:"roles"`
}

// GetUserInfo extracts user information and entitlements from ID token.
func (p *OIDCProvider) GetUserInfo(idToken *oidc.IDToken) (*UserInfo, error) {
	var claims ClaimsWithEntitlements
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}

	entitlements := make([]string, 0, len(claims.Groups)+len(claims.Roles))
	entitlements = append(entitlements, claims.Groups...)
	entitlements = append(entitlements, claims.Roles...)

	return &UserInfo{
		ID:            idToken.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified,
		Name:          claims.Name,
		Picture:       claims.Picture,
		Provider:      "oidc",
		Entitlements:  entitlements,
		Groups:        claims.Groups,
	}, nil
}

// RefreshToken refreshes an expired token.
func (p *OIDCProvider) RefreshToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	tokenSource := p.config.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("refresh token: %w", err)
	}
	return newToken, nil
}
