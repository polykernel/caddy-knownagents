// SPDX-FileCopyrightText: 2024 polykernel
// SPDX-License-Identifier: MIT or Apache-2.0

package caddydarkvisitors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

// The address for the Dark Visitors agent analytics API endpoint.
const AnalyticsEndpoint = "https://api.darkvisitors.com/visits"

// The address for the Dark Visitors robots.txt generation API endpoint.
const RobotsTxtEndpoint = "https://api.darkvisitors.com/robots-txts"

// AgentTypes are groups of agent classified by the Dark Visitors API.
type AgentType = string

const (
	AIAssistant          AgentType = "AI Assistant"
	AIDataScraper        AgentType = "AI Data Scraper"
	AISearchCrawler      AgentType = "AI Search Crawler"
	Archiver             AgentType = "Archiver"
	DeveloperHelper      AgentType = "Developer Helper"
	Fetcher              AgentType = "Fetcher"
	HeadlessBrowser      AgentType = "Headless Browser"
	IntelligenceGatherer AgentType = "Intelligence Gatherer"
	Scraper              AgentType = "Scraper"
	SearchEngineCrawlers AgentType = "Search Engine Crawler"
	SEOCrawler           AgentType = "SEO Crawler"
	Uncategorized        AgentType = "Uncategorized"
	UndocumentedAIAgent  AgentType = "Undocumented AI Agent"
)

// allAgentTypes is a list of all documented Dark Visitors agent types.
var allAgentTypes = []AgentType{
	AIAssistant,
	AIDataScraper,
	AISearchCrawler,
	Archiver,
	DeveloperHelper,
	Fetcher,
	HeadlessBrowser,
	IntelligenceGatherer,
	Scraper,
	SearchEngineCrawlers,
	SEOCrawler,
	Uncategorized,
	UndocumentedAIAgent,
}

func init() {
	caddy.RegisterModule(Darkvisitors{})
	httpcaddyfile.RegisterHandlerDirective("darkvisitors", parseCaddyfile)
	httpcaddyfile.RegisterDirectiveOrder("darkvisitors", httpcaddyfile.Before, "header")
}

// Darkvisitors is a middleware which implements a HTTP handler that sends
// HTTP request information as visit events to the Dark Visitors API.
//
// Its API is still experimental and may be subject to change.
type Darkvisitors struct {
	// The access token used to authenticate to the Dark Visitors agent
	// analytics API endpoint.
	AccessToken string `json:"access_token"`

	// Enables generation of robots.txt derived from agent analytics data using
	// the Dark Visitors robots.txt generation API endpoint.
	RobotsTxt *RobotsTxt `json:"robots_txt,omitempty"`

	logger *zap.Logger
}

// RobotsTxt configures automated generation of robots.txt via the Dark Visitors API.
type RobotsTxt struct {
	// A list of agent types to block.
	AgentTypes []AgentType `json:"agent_types"`

	// The path to disallow access for the specified agent types.
	Disallow string `json:"disallow,omitempty"`

	text string `json:"-"`
}

// CaddyModule returns the Caddy module information.
func (Darkvisitors) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.darkvisitors",
		New: func() caddy.Module { return new(Darkvisitors) },
	}
}

// FetchRobotsTxt queries the Dark Visitors robots.txt generation API endpoint
// and stores the returned robots.txt content.
func (m *Darkvisitors) FetchRobotsTxt(ctx caddy.Context) {
	m.logger.Info("Fetching generated robots.txt")

	query, err := json.Marshal(m.RobotsTxt)
	if err != nil {
		m.logger.Error("Error marshaling robots.txt query", zap.Error(err))
		return
	}

	m.logger.Debug("Robots.txt query payload constructed", zap.ByteString("payload", query))

	client := &http.Client{}
	req, err := http.NewRequest("POST", RobotsTxtEndpoint, bytes.NewBuffer(query))
	if err != nil {
		m.logger.Error("Error creating request", zap.Error(err))
		return
	}

	req.Header.Set("Authorization", "Bearer "+m.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		m.logger.Warn("Error sending robots.txt query", zap.Error(err))
		return
	}
	m.logger.Debug("Robots.txt query sent", zap.Int("status", resp.StatusCode))
	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		m.logger.Warn("Error reading response body", zap.Error(err))
		return
	}
	m.RobotsTxt.text = string(body)
}

// Provision implements caddy.Provisioner.
func (m *Darkvisitors) Provision(ctx caddy.Context) error {
	repl := caddy.NewReplacer()

	m.AccessToken = repl.ReplaceAll(m.AccessToken, "")

	m.logger = ctx.Logger()

	if m.RobotsTxt != nil {
		if m.RobotsTxt.Disallow == "" {
			m.RobotsTxt.Disallow = "/"
		}
		go m.FetchRobotsTxt(ctx)
	}

	return nil
}

// Validate implements caddy.Validator.
func (m Darkvisitors) Validate() error {
	m.logger.Debug("Access Token: " + m.AccessToken)

	if m.RobotsTxt != nil {
		// check if the supplied agent types are valid
		for _, at := range m.RobotsTxt.AgentTypes {
			if !slices.Contains(allAgentTypes, at) {
				return fmt.Errorf("unrecognized agent type '%s'", at)
			}
		}

		m.logger.Debug("Agent Types: " + strings.Join(m.RobotsTxt.AgentTypes, ","))
		m.logger.Debug("Disallow: " + m.RobotsTxt.Disallow)
	}

	m.logger.Info("Darkvisitors middleware validated")

	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m Darkvisitors) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
	next caddyhttp.Handler,
) error {
	if m.RobotsTxt != nil {
		caddyhttp.SetVar(r.Context(), "dv_robots_txt", m.RobotsTxt.text)
	}

	// run the next handler
	err := next.ServeHTTP(w, r)
	if err != nil {
		return err
	}

	go func() {
		sanitizedHeaders := r.Header.Clone()
		sanitizedHeaders.Del("Cookie")

		visitEvent := map[string]interface{}{
			"request_path":    r.URL.Path,
			"request_method":  r.Method,
			"request_headers": sanitizedHeaders,
		}

		body, err := json.Marshal(visitEvent)
		if err != nil {
			m.logger.Error("Error marshaling visitor event", zap.Error(err))
			return
		}

		m.logger.Debug("Visit event payload constructed", zap.ByteString("payload", body))

		client := &http.Client{}
		req, err := http.NewRequest("POST", AnalyticsEndpoint, bytes.NewBuffer(body))
		if err != nil {
			m.logger.Error("Error creating request", zap.Error(err))
			return
		}

		req.Header.Set("Authorization", "Bearer "+m.AccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			m.logger.Warn("Error sending visitor event", zap.Error(err))
		} else {
			m.logger.Debug("Visitor event sent", zap.Int("status", resp.StatusCode))
		}
		defer func() {
			_ = resp.Body.Close()
		}()
	}()

	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *Darkvisitors) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // consume directive name

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "robots_txt":
			if m.RobotsTxt != nil {
				return d.Err("robots_txt is already configured")
			}
			m.RobotsTxt = new(RobotsTxt)
			for nesting := d.Nesting(); d.NextBlock(nesting); {
				switch d.Val() {
				case "agent_types":
					if !d.NextArg() {
						return d.ArgErr()
					}
					if d.Val() == "*" {
						m.RobotsTxt.AgentTypes = allAgentTypes
						if d.NextArg() {
							return d.Errf("unexpected argument '%s'", d.Val())
						}
					} else {
						m.RobotsTxt.AgentTypes = append(m.RobotsTxt.AgentTypes, d.Val())
						for d.NextArg() {
							m.RobotsTxt.AgentTypes = append(m.RobotsTxt.AgentTypes, d.Val())
						}
					}

				case "disallow":
					if !d.NextArg() {
						return d.ArgErr()
					}
					m.RobotsTxt.Disallow = d.Val()
				default:
					return d.Errf("unknown subdirective '%s'", d.Val())
				}
			}
		case "access_token":
			if !d.NextArg() {
				return d.ArgErr()
			}
			m.AccessToken = d.Val()
		default:
			return d.Errf("unrecognized subdirective '%s'", d.Val())
		}
	}

	if d.NextArg() {
		return d.Errf("unexpected argument '%s'", d.Val())
	}

	if m.AccessToken == "" {
		return d.Err("missing access token")
	}

	if m.RobotsTxt != nil {
		if len(m.RobotsTxt.AgentTypes) == 0 {
			return d.Err("missing agent type filters")
		}
	}

	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Darkvisitors middleware.
//
// Syntax:
//
//	darkvisitors {
//	  access_token <token>
//	}
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m Darkvisitors
	err := m.UnmarshalCaddyfile(h.Dispenser)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Darkvisitors)(nil)
	_ caddy.Validator             = (*Darkvisitors)(nil)
	_ caddyhttp.MiddlewareHandler = (*Darkvisitors)(nil)
	_ caddyfile.Unmarshaler       = (*Darkvisitors)(nil)
)
