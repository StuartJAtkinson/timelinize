package mcp_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/StuartJAtkinson/timelinize/mcp"
)

func TestMCPServer_EndToEnd(t *testing.T) {
	var (
		lastMethod string
		lastPath   string
		lastQuery  string
		lastOrigin string
		lastBody   string
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lastMethod = r.Method
		lastPath = r.URL.Path
		lastQuery = r.URL.RawQuery
		lastOrigin = r.Header.Get("Origin")
		b, _ := io.ReadAll(r.Body)
		lastBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	runRPC := func(t *testing.T, baseURL, input string) []map[string]any {
		t.Helper()
		var out bytes.Buffer
		if err := mcp.Serve(context.Background(), strings.NewReader(input), &out, baseURL); err != nil {
			t.Fatalf("Serve error: %v", err)
		}
		var resps []map[string]any
		for _, line := range strings.Split(strings.TrimSpace(out.String()), "\n") {
			if line == "" {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal([]byte(line), &m); err != nil {
				t.Fatalf("unmarshal response %q: %v", line, err)
			}
			resps = append(resps, m)
		}
		return resps
	}

	// 1. Initialize + notification (no response for notification) + tools/list + resources/read
	resps := runRPC(t, ts.URL, strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05"}}`,
		`{"jsonrpc":"2.0","method":"notifications/initialized"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":3,"method":"resources/read","params":{"uri":"tlz://repositories"}}`,
	}, "\n")+"\n")

	if len(resps) != 3 {
		t.Fatalf("want 3 responses (notification skipped), got %d", len(resps))
	}

	// Verify exact 11 allowed tools and absence of forbidden endpoints.
	toolsResult := resps[1]["result"].(map[string]any)
	tools := toolsResult["tools"].([]any)
	if len(tools) != 11 {
		t.Fatalf("want 11 tools, got %d", len(tools))
	}
	forbidden := map[string]bool{
		"delete-items": true, "delete_items": true,
		"merge-entities": true, "merge_entities": true,
		"import": true, "plan-import": true, "plan_import": true,
		"file-stat": true, "file_stat": true,
		"file-listing": true, "file_listing": true,
		"file-selector-roots": true, "file_selector_roots": true,
		"logs": true,
	}
	for _, raw := range tools {
		name := raw.(map[string]any)["name"].(string)
		if forbidden[name] {
			t.Fatalf("forbidden tool exposed: %s", name)
		}
	}

	// Verify resource read hit /api/open-repositories with matching Origin header.
	if lastMethod != http.MethodGet || lastPath != "/api/open-repositories" || lastOrigin != ts.URL {
		t.Fatalf("unexpected resource request: method=%s path=%s origin=%s", lastMethod, lastPath, lastOrigin)
	}

	// 2. Verify search_items caps limit at MaxPageSize (50).
	_ = runRPC(t, ts.URL, `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"search_items","arguments":{"repo":"r1","limit":5000}}}`+"\n")
	if lastMethod != http.MethodPost || lastPath != "/api/search-items" {
		t.Fatalf("unexpected search_items request: %s %s", lastMethod, lastPath)
	}
	var searchPayload map[string]any
	if err := json.Unmarshal([]byte(lastBody), &searchPayload); err != nil {
		t.Fatal(err)
	}
	if int(searchPayload["limit"].(float64)) != mcp.MaxPageSize {
		t.Fatalf("want capped limit %d, got %v", mcp.MaxPageSize, searchPayload["limit"])
	}

	// 3. Verify item_classifications sends a JSON string payload and jobs uses QUERY method.
	_ = runRPC(t, ts.URL, `{"jsonrpc":"2.0","id":5,"method":"tools/call","params":{"name":"item_classifications","arguments":{"repo_id":"repo-123"}}}`+"\n")
	if lastPath != "/api/item-classifications" || strings.TrimSpace(lastBody) != `"repo-123"` {
		t.Fatalf("want string body \"repo-123\" for item-classifications, got path=%s body=%s", lastPath, lastBody)
	}

	_ = runRPC(t, ts.URL, `{"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"jobs","arguments":{"repo_id":"repo-123","most_recent":5}}}`+"\n")
	if lastMethod != "QUERY" || lastPath != "/api/jobs" {
		t.Fatalf("want QUERY /api/jobs, got %s %s", lastMethod, lastPath)
	}

	_ = runRPC(t, ts.URL, `{"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"charts","arguments":{"repo_id":"repo-123","name":"periodical","period":"monthly"}}}`+"\n")
	if lastMethod != http.MethodGet || lastPath != "/api/charts" || !strings.Contains(lastQuery, "name=periodical") {
		t.Fatalf("unexpected charts request: %s %s?%s", lastMethod, lastPath, lastQuery)
	}

	// 4. Verify offline server returns the exact required sentence from MCP.md.
	offlineResps := runRPC(t, "http://127.0.0.1:1", `{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"build_info"}}`+"\n")
	callRes := offlineResps[0]["result"].(map[string]any)
	if callRes["isError"] != true {
		t.Fatalf("want isError=true when app is not running, got %v", callRes)
	}
	content := callRes["content"].([]any)[0].(map[string]any)["text"].(string)
	wantMsg := "The app must already be running (http://127.0.0.1:12002)."
	if content != wantMsg {
		t.Fatalf("want %q, got %q", wantMsg, content)
	}
}
