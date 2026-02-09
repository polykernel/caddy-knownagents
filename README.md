# Caddy Knownagents Module

[![Go Reference](https://pkg.go.dev/badge/github.com/polykernel/caddy-knownagents.svg)](https://pkg.go.dev/github.com/polykernel/caddy-knownagents)

A super simple [Caddy](https://caddyserver.com/) module for interacting with the [Known Agents API](https://knownagents.com/docs/analytics).

## Building

To compile this Caddy module, follow the instructions from [Build from Source](https://caddyserver.com/docs/build) and import the `github.com/polykernel/caddy-knownagents` module.

## Configuration

### Syntax

```Caddyfile
knownagents {
  access_token <token>
  robots_txt {
    agent_types <types...>
    disallow <path>
  }
}
```

- **access_token** sets the OAuth authorization token used to communicate with the Known Agents API. Global placeholders are supported in the argument.
- **robots_txt** enables generation of robots.txt derived from agent analytics data using the Known Agents API.
  - **agent_types** specifies a list of [agent types](https://knownagents.com/agents) to be blocked by the generated robots.txt. The special token "\*" is supported as an argument which resolves to all documented agent types. Note: when "\*" is passed, there must be no further arguments.
  - **disallow** specifies the path to disallow for the specified agent types. Default: `/`.

If the `robots_txt` block is configured, then the special variable `http.vars.ka_robots_txt` in the HTTP request context will be set to the raw content of the robots.txt returned by the Known Agents API. Note: the robots.txt query is performed once during the provision phase of the module lifecycle and cached thereafter.

By default, the `knownagents` directive is ordered before [`header`](https://caddyserver.com/docs/caddyfile/directives#directive-header) in the Caddyfile. This ensures that the raw request content (sensitive data such as cookies are still stripped) is used to build a visit event. If this order does not fit your needs, you can change the order using the global [`order`](https://caddyserver.com/docs/caddyfile/directives#directive-order) directive. For example:

```Caddyfile
{
  order knownagents before handle
}
```

### Example

A basic Caddyfile configuration is provided below:

```Caddyfile
knownagents {
  access_token {env.KA_ACCESS_TOKEN}
  robots_txt {
    agent_types "AI Assistant" "AI Data Scraper"
    disallow /
  }
}
```

## License

Copyright (c) 2024 polykernel

The source code in this repository is made available under the [MIT](https://spdx.org/licenses/MIT.html) or [Apache 2.0](https://spdx.org/licenses/Apache-2.0.html) license.
