// Package mcp implements the stdio Model Context Protocol (MCP) server for Timelinize,
// wrapping the local read-only HTTP API at http://127.0.0.1:12002/api/ per MCP.md.
package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the local Timelinize admin server address.
	DefaultBaseURL = "http://127.0.0.1:12002"

	// MaxPageSize caps search/conversation page sizes so broad queries cannot dump an entire timeline into a context window.
	MaxPageSize = 50

	notRunningMsg = "The app must already be running (http://127.0.0.1:12002)."
)

// Server serves MCP JSON-RPC 2.0 requests over stdio against a running Timelinize HTTP API.
type Server struct {
	BaseURL string
	Client  *http.Client
}

// Serve runs the stdio MCP loop until r reaches EOF or ctx is canceled.
func Serve(ctx context.Context, r io.Reader, w io.Writer, baseURL string) error {
	s := &Server{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
	if s.BaseURL == "" {
		s.BaseURL = DefaultBaseURL
	}
	return s.Run(ctx, r, w)
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Run processes newline-delimited JSON-RPC 2.0 messages from r and writes responses to w.
func (s *Server) Run(ctx context.Context, r io.Reader, w io.Writer) error {
	if s.BaseURL == "" {
		s.BaseURL = DefaultBaseURL
	}
	if s.Client == nil {
		s.Client = &http.Client{Timeout: 30 * time.Second}
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	enc := json.NewEncoder(w)

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      json.RawMessage("null"),
				Error:   &rpcError{Code: -32700, Message: "invalid JSON"},
			})
			continue
		}

		// Notifications have no ID and receive no response.
		if len(req.ID) == 0 {
			continue
		}

		res, rpcErr := s.dispatch(ctx, req.Method, req.Params)
		resp := rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
		}
		if rpcErr != nil {
			resp.Error = rpcErr
		} else {
			resp.Result = res
		}
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) dispatch(ctx context.Context, method string, rawParams json.RawMessage) (any, *rpcError) {
	switch method {
	case "initialize":
		version := "2024-11-05"
		var p struct {
			ProtocolVersion string `json:"protocolVersion"`
		}
		if len(rawParams) > 0 && json.Unmarshal(rawParams, &p) == nil && p.ProtocolVersion != "" {
			version = p.ProtocolVersion
		}
		return map[string]any{
			"protocolVersion": version,
			"capabilities": map[string]any{
				"tools":     map[string]any{},
				"resources": map[string]any{},
			},
			"serverInfo": map[string]any{
				"name":    "timelinize",
				"version": "0.1.0",
			},
		}, nil

	case "ping":
		return map[string]any{}, nil

	case "tools/list":
		return map[string]any{"tools": toolDefinitions}, nil

	case "tools/call":
		var call struct {
			Name      string         `json:"name"`
			Arguments map[string]any `json:"arguments"`
		}
		if err := json.Unmarshal(rawParams, &call); err != nil {
			return nil, &rpcError{Code: -32602, Message: "invalid tool call arguments"}
		}
		text, isErr, rpcErr := s.callTool(ctx, call.Name, call.Arguments)
		if rpcErr != nil {
			return nil, rpcErr
		}
		return map[string]any{
			"content": []map[string]any{{"type": "text", "text": text}},
			"isError": isErr,
		}, nil

	case "resources/list":
		return map[string]any{"resources": resourceDefinitions}, nil

	case "resources/read":
		var p struct {
			URI string `json:"uri"`
		}
		if err := json.Unmarshal(rawParams, &p); err != nil || p.URI == "" {
			return nil, &rpcError{Code: -32602, Message: "resource uri is required"}
		}
		text, rpcErr := s.readResource(ctx, p.URI)
		if rpcErr != nil {
			return nil, rpcErr
		}
		return map[string]any{
			"contents": []map[string]any{{
				"uri":      p.URI,
				"mimeType": "application/json",
				"text":     text,
			}},
		}, nil

	default:
		return nil, &rpcError{Code: -32601, Message: "method not found: " + method}
	}
}

func (s *Server) callTool(ctx context.Context, name string, args map[string]any) (string, bool, *rpcError) {
	if args == nil {
		args = map[string]any{}
	}

	var (
		method   string
		endpoint string
		payload  any
	)

	switch name {
	case "search_items":
		method, endpoint, payload = http.MethodPost, "search-items", capLimit(args)
	case "search_entities":
		method, endpoint, payload = http.MethodPost, "search-entities", capLimit(args)
	case "get_entity":
		method, endpoint, payload = http.MethodPost, "get-entity", map[string]any{
			"repo_id":   args["repo_id"],
			"entity_id": args["entity_id"],
		}
	case "conversation":
		method, endpoint, payload = http.MethodPost, "conversation", capLimit(args)
	case "recent_conversations":
		method, endpoint, payload = http.MethodPost, "recent-conversations", capLimit(args)
	case "list_repositories":
		method, endpoint = http.MethodGet, "open-repositories"
	case "list_data_sources":
		method, endpoint = http.MethodGet, "data-sources"
	case "item_classifications":
		repoID, _ := args["repo_id"].(string)
		if repoID == "" {
			repoID, _ = args["repo"].(string)
		}
		method, endpoint, payload = http.MethodPost, "item-classifications", repoID
	case "charts":
		q := url.Values{}
		for _, k := range []string{"name", "repo_id", "period", "days", "entity_id"} {
			if v, ok := args[k]; ok && v != nil {
				q.Set(k, fmt.Sprint(v))
			}
		}
		method, endpoint = http.MethodGet, "charts?"+q.Encode()
	case "jobs":
		method, endpoint, payload = "QUERY", "jobs", map[string]any{
			"repo_id":     args["repo_id"],
			"job_ids":     args["job_ids"],
			"most_recent": args["most_recent"],
		}
	case "build_info":
		method, endpoint = http.MethodGet, "build-info"
	default:
		return "", false, &rpcError{Code: -32602, Message: "unknown tool: " + name}
	}

	body, err := s.request(ctx, method, endpoint, payload)
	if err != nil {
		return err.Error(), true, nil
	}
	return body, false, nil
}

func (s *Server) readResource(ctx context.Context, uri string) (string, *rpcError) {
	var endpoint string
	switch uri {
	case "tlz://repositories":
		endpoint = "open-repositories"
	case "tlz://data-sources":
		endpoint = "data-sources"
	default:
		return "", &rpcError{Code: -32602, Message: "unknown resource: " + uri}
	}
	body, err := s.request(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", &rpcError{Code: -32603, Message: err.Error()}
	}
	return body, nil
}

func capLimit(args map[string]any) map[string]any {
	out := make(map[string]any, len(args)+1)
	for k, v := range args {
		out[k] = v
	}
	limit := 0
	switch v := out["limit"].(type) {
	case float64:
		limit = int(v)
	case int:
		limit = v
	case string:
		limit, _ = strconv.Atoi(v)
	}
	if limit <= 0 || limit > MaxPageSize {
		out["limit"] = MaxPageSize
	}
	return out
}

func (s *Server) request(ctx context.Context, method, endpoint string, payload any) (string, error) {
	var bodyReader io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		bodyReader = bytes.NewReader(b)
	}

	reqURL := s.BaseURL + "/api/" + endpoint
	req, err := http.NewRequestWithContext(ctx, method, reqURL, bodyReader)
	if err != nil {
		return "", err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Origin", s.BaseURL)

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", errors.New(notRunningMsg)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		msg := strings.TrimSpace(string(respBytes))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}
	return strings.TrimSpace(string(respBytes)), nil
}

var resourceDefinitions = []map[string]any{
	{
		"uri":         "tlz://repositories",
		"name":        "Open repositories",
		"description": "Currently open Timelinize repositories",
		"mimeType":    "application/json",
	},
	{
		"uri":         "tlz://data-sources",
		"name":        "Data sources",
		"description": "Available data source importers",
		"mimeType":    "application/json",
	},
}

var toolDefinitions = []map[string]any{
	{
		"name":        "search_items",
		"description": "Search items in an open timeline (page size capped at 50).",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo":            map[string]any{"type": "string", "description": "Open repository UUID"},
				"data_text":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"classification":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"data_source":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"entity_id":       map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
				"start_timestamp": map[string]any{"type": "string", "description": "RFC3339 start timestamp"},
				"end_timestamp":   map[string]any{"type": "string", "description": "RFC3339 end timestamp"},
				"limit":           map[string]any{"type": "integer", "description": "Max items to return (capped at 50)"},
				"offset":          map[string]any{"type": "integer"},
				"only_total":      map[string]any{"type": "boolean", "description": "Return only the matching count"},
			},
			"required": []string{"repo"},
		},
	},
	{
		"name":        "search_entities",
		"description": "Search people, places, or accounts in an open timeline (page size capped at 50).",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo":     map[string]any{"type": "string", "description": "Open repository UUID"},
				"name":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"type":     map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
				"order_by": map[string]any{"type": "string"},
				"limit":    map[string]any{"type": "integer", "description": "Max entities to return (capped at 50)"},
			},
			"required": []string{"repo"},
		},
	},
	{
		"name":        "get_entity",
		"description": "Get a single entity by ID.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo_id":   map[string]any{"type": "string"},
				"entity_id": map[string]any{"type": "integer"},
			},
			"required": []string{"repo_id", "entity_id"},
		},
	},
	{
		"name":        "conversation",
		"description": "Load messages in a conversation matching the given filter (page size capped at 50).",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo":      map[string]any{"type": "string"},
				"entity_id": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
				"limit":     map[string]any{"type": "integer"},
				"offset":    map[string]any{"type": "integer"},
			},
			"required": []string{"repo"},
		},
	},
	{
		"name":        "recent_conversations",
		"description": "List recent conversation threads in a timeline (page size capped at 50).",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo":      map[string]any{"type": "string"},
				"entity_id": map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
				"limit":     map[string]any{"type": "integer"},
			},
			"required": []string{"repo"},
		},
	},
	{
		"name":        "list_repositories",
		"description": "List currently open timeline repositories.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
	},
	{
		"name":        "list_data_sources",
		"description": "List data source importers supported by this build.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
	},
	{
		"name":        "item_classifications",
		"description": "List the item classification vocabulary for a timeline.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo_id": map[string]any{"type": "string"},
			},
			"required": []string{"repo_id"},
		},
	},
	{
		"name":        "charts",
		"description": "Return timeline statistics for a named chart (periodical, classifications, datasources, attributes_stacked_area, recent_data_sources).",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo_id":   map[string]any{"type": "string"},
				"name":      map[string]any{"type": "string"},
				"period":    map[string]any{"type": "string", "description": "monthly, daily, or yearly (for periodical)"},
				"days":      map[string]any{"type": "integer", "description": "Number of days (for recent_data_sources)"},
				"entity_id": map[string]any{"type": "integer", "description": "Entity ID (for attributes_stacked_area)"},
			},
			"required": []string{"repo_id", "name"},
		},
	},
	{
		"name":        "jobs",
		"description": "Read-only job status for a repository.",
		"inputSchema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo_id":     map[string]any{"type": "string"},
				"job_ids":     map[string]any{"type": "array", "items": map[string]any{"type": "integer"}},
				"most_recent": map[string]any{"type": "integer"},
			},
			"required": []string{"repo_id"},
		},
	},
	{
		"name":        "build_info",
		"description": "Return build OS and architecture info.",
		"inputSchema": map[string]any{"type": "object", "properties": map[string]any{}},
	},
}
