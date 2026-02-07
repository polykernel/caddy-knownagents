// SPDX-FileCopyrightText: 2024 polykernel
// SPDX-License-Identifier: MIT or Apache-2.0

package caddyknownagents

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
)

func TestUnmarshalModule(t *testing.T) {
	fmt.Println("Testing unmarshal module... ")
	access_token := "aHVudGVyMg=="
	config := fmt.Sprintf(`knownagents {
		access_token %s
	}`, access_token)

	dispenser := caddyfile.NewTestDispenser(config)
	m := &Knownagents{}

	err := m.UnmarshalCaddyfile(dispenser)
	if err != nil {
		t.Errorf("UnmarshalCaddyfile failed with %v", err)
		return
	}

	expected := access_token
	got := m.AccessToken
	if expected != got {
		t.Errorf("Expected AccessToken to be '%s' but got '%s'", expected, got)
		return
	}
}

func TestUnmarshalModuleRobotsTxt(t *testing.T) {
	fmt.Println("Testing unmarshal module (robots.txt block)... ")
	access_token := "aHVudGVyMg=="
	agent_types := []string{AIAssistant, AISearchCrawler, AIDataScraper}
	var agent_types_str string
	{
		var b strings.Builder
		for i := 0; i < len(agent_types); i++ {
			fmt.Fprintf(&b, "%q", agent_types[i])
		}
		agent_types_str = b.String()
	}
	disallow := "/"
	config := fmt.Sprintf(`knownagents {
		access_token %s
		robots_txt {
			agent_types %s
			disallow %s
		}
	}`, access_token, agent_types_str, disallow)

	dispenser := caddyfile.NewTestDispenser(config)
	m := &Knownagents{}

	err := m.UnmarshalCaddyfile(dispenser)
	if err != nil {
		t.Errorf("UnmarshalCaddyfile failed with %v", err)
		return
	}

	{
		expected := access_token
		got := m.AccessToken
		if expected != got {
			t.Errorf("Expected AccessToken to be '%s' but got '%s'", expected, got)
			return
		}
	}

	if m.RobotsTxt == nil {
		t.Error("Expected RobotsTxt not to be nil")
		return
	}

	{
		expected := agent_types
		got := m.RobotsTxt.AgentTypes
		if !slices.Equal(expected, got) {
			t.Errorf("Expected AccessToken to be '%s' but got '%s'", expected, got)
			return
		}
	}

	{
		expected := disallow
		got := m.RobotsTxt.Disallow
		if expected != got {
			t.Errorf("Expected Disallow to be '%s' but got '%s'", expected, got)
			return
		}
	}
}

func TestUnmarshalModuleRobotsTxtWildcard(t *testing.T) {
	fmt.Println("Testing unmarshal module (robots.txt block + wildcard)... ")
	access_token := "aHVudGVyMg=="
	disallow := "/"
	config := fmt.Sprintf(`knownagents {
		access_token %s
		robots_txt {
			agent_types *
			disallow %s
		}
	}`, access_token, disallow)

	dispenser := caddyfile.NewTestDispenser(config)
	m := &Knownagents{}

	err := m.UnmarshalCaddyfile(dispenser)
	if err != nil {
		t.Errorf("UnmarshalCaddyfile failed with %v", err)
		return
	}

	{
		expected := access_token
		got := m.AccessToken
		if expected != got {
			t.Errorf("Expected AccessToken to be '%s' but got '%s'", expected, got)
			return
		}
	}

	if m.RobotsTxt == nil {
		t.Error("Expected RobotsTxt not to be nil")
		return
	}

	{
		expected := allAgentTypes
		got := m.RobotsTxt.AgentTypes
		if !slices.Equal(expected, got) {
			t.Errorf("Expected AccessToken to be '%s' but got '%s'", expected, got)
			return
		}
	}

	{
		expected := disallow
		got := m.RobotsTxt.Disallow
		if expected != got {
			t.Errorf("Expected Disallow to be '%s' but got '%s'", expected, got)
			return
		}
	}
}
