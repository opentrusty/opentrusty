package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TestMetadata holds info parsed from Go source comments
type TestMetadata struct {
	Name        string `json:"name"`
	Purpose     string `json:"purpose,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Security    string `json:"security,omitempty"`
	Permissions string `json:"permissions,omitempty"`
	Expected    string `json:"expected,omitempty"`
	TestCaseID  string `json:"test_case_id,omitempty"`
	Package     string `json:"package"`
	Category    string `json:"category"`
	Type        string `json:"type"` // UT, ST, E2E, etc.
}

// GoTestEvent represents a single event from 'go test -json'
type GoTestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

// FinalTestResult is the merged result for a single test
type FinalTestResult struct {
	Name        string       `json:"name"`
	Status      string       `json:"status"`
	Elapsed     float64      `json:"elapsed_seconds"`
	Package     string       `json:"package"`
	Failure     string       `json:"failure_reason,omitempty"`
	Annotations TestMetadata `json:"annotations"`
}

// ReportSummary holds top-level stats
type ReportSummary struct {
	GeneratedAt time.Time         `json:"generated_at"`
	Total       int               `json:"total"`
	Passed      int               `json:"passed"`
	Failed      int               `json:"failed"`
	Skipped     int               `json:"skipped"`
	Results     []FinalTestResult `json:"results"`
}

func main() {
	inputPath := flag.String("input", "", "Path to go test -json output file")
	outputJSON := flag.String("out-json", "", "Path for output JSON report")
	outputMD := flag.String("out-md", "", "Path for output Markdown report")
	outputHTML := flag.String("out-html", "", "Path for output HTML report")
	title := flag.String("title", "Test Report", "Report title")
	filterCats := flag.String("filter-categories", "", "Comma-separated list of categories to include")
	excludeCats := flag.String("exclude-categories", "", "Comma-separated list of categories to exclude")
	filterType := flag.String("filter-type", "", "Filter by test type (UT, ST, E2E, etc.)")
	excludeType := flag.String("exclude-type", "", "Exclude by test type (UT, ST, E2E, etc.)")
	flag.Parse()

	if *inputPath == "" || *outputJSON == "" || *outputMD == "" {
		fmt.Println("Usage: report_gen -input <json_file> -out-json <out_json> -out-md <out_md>")
		os.Exit(1)
	}

	// 1. Scan codebase for annotations
	metadataMap := scanMetadata()

	// 2. Parse go test -json output
	results := parseTestOutput(*inputPath, metadataMap)

	// 3. Filter Results if requested
	if *filterCats != "" {
		cats := strings.Split(*filterCats, ",")
		filtered := []FinalTestResult{}
		for _, res := range results {
			for _, cat := range cats {
				if strings.TrimSpace(cat) == res.Annotations.Category {
					filtered = append(filtered, res)
					break
				}
			}
		}
		results = filtered
	}

	if *excludeCats != "" {
		cats := strings.Split(*excludeCats, ",")
		filtered := []FinalTestResult{}
		for _, res := range results {
			excluded := false
			for _, cat := range cats {
				if strings.TrimSpace(cat) == res.Annotations.Category {
					excluded = true
					break
				}
			}
			if !excluded {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	}

	if *filterType != "" {
		filtered := []FinalTestResult{}
		for _, res := range results {
			if strings.EqualFold(res.Annotations.Type, *filterType) {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	}

	if *excludeType != "" {
		filtered := []FinalTestResult{}
		for _, res := range results {
			if !strings.EqualFold(res.Annotations.Type, *excludeType) {
				filtered = append(filtered, res)
			}
		}
		results = filtered
	}

	// 4. Generate Reports
	summary := generateSummary(results)

	// 4. Save JSON
	saveJSON(summary, *outputJSON)

	// 5. Save Markdown
	saveMarkdown(summary, *outputMD, *title)

	// 6. Save HTML
	if *outputHTML != "" {
		saveHTML(summary, *outputHTML, *title)
	}

	// 7. Exit with error if any tests failed to ensure CI gates work correctly
	if summary.Failed > 0 {
		fmt.Printf("\n❌ Test Reporting: %d tests failed. Exiting with error.\n", summary.Failed)
		os.Exit(1)
	}
}

func scanMetadata() map[string]TestMetadata {
	metadataMap := make(map[string]TestMetadata)
	fset := token.NewFileSet()

	filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		if strings.Contains(path, "vendor/") || strings.Contains(path, ".git/") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		pkgPath := getPackagePath(path)

		for _, decl := range node.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(fn.Name.Name, "Test") {
				continue
			}

			meta := TestMetadata{
				Name:     fn.Name.Name,
				Package:  pkgPath,
				Type:     determineType(pkgPath),
				Category: determineCategory(pkgPath, fn.Name.Name),
			}

			if fn.Doc != nil {
				for _, line := range fn.Doc.List {
					text := strings.TrimSpace(strings.TrimPrefix(line.Text, "//"))
					if strings.HasPrefix(text, "TestPurpose:") {
						meta.Purpose = strings.TrimSpace(strings.TrimPrefix(text, "TestPurpose:"))
					} else if strings.HasPrefix(text, "Scope:") {
						meta.Scope = strings.TrimSpace(strings.TrimPrefix(text, "Scope:"))
					} else if strings.HasPrefix(text, "Security:") {
						meta.Security = strings.TrimSpace(strings.TrimPrefix(text, "Security:"))
					} else if strings.HasPrefix(text, "Permissions:") {
						meta.Permissions = strings.TrimSpace(strings.TrimPrefix(text, "Permissions:"))
					} else if strings.HasPrefix(text, "Expected:") {
						meta.Expected = strings.TrimSpace(strings.TrimPrefix(text, "Expected:"))
					} else if strings.HasPrefix(text, "Test Case ID:") {
						meta.TestCaseID = strings.TrimSpace(strings.TrimPrefix(text, "Test Case ID:"))
					}
				}
			}
			key := fmt.Sprintf("%s.%s", pkgPath, fn.Name.Name)
			metadataMap[key] = meta
		}
		return nil
	})

	return metadataMap
}

func getPackagePath(filePath string) string {
	dir := filepath.Dir(filePath)
	// Relative to repo root
	dir = strings.TrimPrefix(dir, "./")
	if dir == "." {
		return "main"
	}
	// Convert to module path (simplified)
	return "github.com/opentrusty/opentrusty/" + dir
}

func determineType(pkgPath string) string {
	// Root module prefix
	const prefix = "github.com/opentrusty/opentrusty/"
	relPath := strings.TrimPrefix(pkgPath, prefix)

	if strings.HasPrefix(relPath, "tests/") {
		parts := strings.Split(relPath, "/")
		if len(parts) > 1 {
			return strings.ToUpper(parts[1])
		}
	}
	return "UT"
}

func determineCategory(pkgPath, testName string) string {
	if strings.Contains(pkgPath, "authz") {
		return "AuthZ"
	}
	if strings.Contains(pkgPath, "identity") {
		return "AuthN"
	}
	if strings.Contains(pkgPath, "tenant") {
		return "Tenant"
	}
	if strings.Contains(pkgPath, "oidc") {
		return "OIDC"
	}
	if strings.Contains(pkgPath, "oauth2") {
		return "OAuth2"
	}
	if strings.Contains(pkgPath, "audit") {
		return "Audit"
	}
	if strings.Contains(pkgPath, "transport/http") {
		if strings.Contains(testName, "Auth") {
			return "Auth API"
		}
		if strings.Contains(testName, "Tenant") {
			return "Tenant API"
		}
		return "API"
	}
	t := determineType(pkgPath)
	if t != "UT" {
		return t + " Tests"
	}
	return "Other"
}

func parseTestOutput(path string, meta map[string]TestMetadata) []FinalTestResult {
	// Initialize with all known tests from metadata
	testStates := make(map[string]*FinalTestResult)
	for key, m := range meta {
		testStates[key] = &FinalTestResult{
			Name:        m.Name,
			Package:     m.Package,
			Status:      "not run",
			Annotations: m,
		}
	}

	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("Error opening test output: %v\n", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var event GoTestEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		if event.Test == "" {
			continue
		}

		key := fmt.Sprintf("%s.%s", event.Package, event.Test)
		res, ok := testStates[key]
		if !ok {
			// Check if it's a subtest (e.g. TestParent/Sub)
			if strings.Contains(event.Test, "/") {
				parentName := strings.Split(event.Test, "/")[0]
				parentKey := fmt.Sprintf("%s.%s", event.Package, parentName)
				if parentMeta, found := meta[parentKey]; found {
					res = &FinalTestResult{
						Name:    event.Test,
						Package: event.Package,
						Annotations: TestMetadata{
							Name:        event.Test,
							Package:     event.Package,
							Category:    parentMeta.Category,
							Type:        parentMeta.Type,
							Purpose:     parentMeta.Purpose + " (Subtest: " + event.Test + ")",
							Scope:       parentMeta.Scope,
							Security:    parentMeta.Security,
							Expected:    parentMeta.Expected,
							Permissions: parentMeta.Permissions,
							TestCaseID:  parentMeta.TestCaseID,
						},
					}
					testStates[key] = res
				} else {
					res = &FinalTestResult{
						Name:    event.Test,
						Package: event.Package,
						Annotations: TestMetadata{
							Name:     event.Test,
							Package:  event.Package,
							Type:     determineType(event.Package),
							Category: "Other",
						},
					}
					testStates[key] = res
				}
			} else {
				res = &FinalTestResult{
					Name:    event.Test,
					Package: event.Package,
					Annotations: TestMetadata{
						Name:     event.Test,
						Package:  event.Package,
						Type:     determineType(event.Package),
						Category: "Other",
					},
				}
				testStates[key] = res
			}
		}

		switch event.Action {
		case "pass":
			res.Status = "pass"
			res.Elapsed = event.Elapsed
		case "fail":
			res.Status = "fail"
			res.Elapsed = event.Elapsed
		case "skip":
			res.Status = "skip"
		case "output":
			if res.Status == "fail" || res.Status == "" {
				res.Failure += event.Output
			}
		}
	}

	var list []FinalTestResult
	for _, v := range testStates {
		list = append(list, *v)
	}
	return list
}

func generateSummary(results []FinalTestResult) ReportSummary {
	summary := ReportSummary{
		GeneratedAt: time.Now(),
		Results:     results,
	}

	for _, r := range results {
		summary.Total++
		switch r.Status {
		case "pass":
			summary.Passed++
		case "fail":
			summary.Failed++
		case "skip":
			summary.Skipped++
		}
	}

	return summary
}

func saveJSON(summary ReportSummary, path string) {
	data, _ := json.MarshalIndent(summary, "", "  ")
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, data, 0644)
}

func saveMarkdown(summary ReportSummary, path string, title string) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# OpenTrusty %s\n\n", title))
	sb.WriteString(fmt.Sprintf("**Generated:** %s  \n", summary.GeneratedAt.Format("2006-01-02 15:04:05 MST")))

	status := "✅ PASSED"
	if summary.Failed > 0 {
		status = "❌ FAILED"
	}
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", status))

	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Total | Passed | Failed | Skipped | Pass Rate |\n")
	sb.WriteString("|-------|--------|--------|---------|-----------|\n")
	rate := 0.0
	if summary.Total > 0 {
		rate = float64(summary.Passed) / float64(summary.Total) * 100
	}
	sb.WriteString(fmt.Sprintf("| %d | %d | %d | %d | %.1f%% |\n\n", summary.Total, summary.Passed, summary.Failed, summary.Skipped, rate))

	// Group by category
	categories := make(map[string][]FinalTestResult)
	for _, r := range summary.Results {
		cat := r.Annotations.Category
		if cat == "" {
			cat = "Uncategorized"
		}
		categories[cat] = append(categories[cat], r)
	}

	sb.WriteString("## Test Results by Category\n\n")

	// Fixed order for categories
	order := []string{"AuthN", "AuthZ", "Tenant", "OAuth2", "OIDC", "Audit", "Auth API", "Tenant API", "API", "SYSTEM Tests", "E2E Tests", "Other", "Uncategorized"}
	for _, cat := range order {
		tests, ok := categories[cat]
		if !ok || len(tests) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", cat))
		sb.WriteString("| ID | Test Name | Status | Purpose | Security |\n")
		sb.WriteString("|----|-----------|--------|---------|----------|\n")
		for _, t := range tests {
			statusIcon := "✅"
			if t.Status == "fail" {
				statusIcon = "❌"
			}
			if t.Status == "skip" {
				statusIcon = "⏭️"
			}
			if t.Status == "not run" {
				statusIcon = "⚪"
			}

			securityHighlight := t.Annotations.Security
			if securityHighlight != "" {
				securityHighlight = "**" + securityHighlight + "**"
			}

			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				t.Annotations.TestCaseID, t.Name, statusIcon, t.Annotations.Purpose, securityHighlight))
		}
		sb.WriteString("\n")
	}

	if summary.Failed > 0 {
		sb.WriteString("## Failure Details\n\n")
		for _, t := range summary.Results {
			if t.Status == "fail" {
				sb.WriteString(fmt.Sprintf("### %s (%s)\n", t.Name, t.Package))
				sb.WriteString("```\n")
				sb.WriteString(t.Failure)
				sb.WriteString("\n```\n\n")
			}
		}
	}

	sb.WriteString("---\n*Report generated by OpenTrusty Test Infrastructure*\n")

	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(sb.String()), 0644)
}

func saveHTML(summary ReportSummary, path string, title string) {
	statusClass := "status-pass"
	statusText := "PASSED"
	if summary.Failed > 0 {
		statusClass = "status-fail"
		statusText = "FAILED"
	}

	rate := 0.0
	if summary.Total > 0 {
		rate = float64(summary.Passed) / float64(summary.Total) * 100
	}

	var sb strings.Builder
	sb.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>OpenTrusty - ` + title + `</title>
    <style>
        :root {
            --primary: #2563eb;
            --success: #10b981;
            --danger: #ef4444;
            --warning: #f59e0b;
            --bg: #f8fafc;
            --text: #1e293b;
            --border: #e2e8f0;
        }
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: var(--bg); color: var(--text); line-height: 1.5; margin: 0; padding: 2rem; }
        .container { max-width: 1000px; margin: 0 auto; background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        h1 { margin-top: 0; border-bottom: 2px solid var(--border); padding-bottom: 0.5rem; }
        .meta { color: #64748b; margin-bottom: 2rem; }
        .status-badge { display: inline-block; padding: 0.25rem 0.75rem; border-radius: 9999px; font-weight: 600; font-size: 0.875rem; }
        .status-pass { background: #dcfce7; color: #166534; }
        .status-fail { background: #fee2e2; color: #991b1b; }
        .summary-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 1rem; margin-bottom: 2rem; }
        .summary-card { background: var(--bg); padding: 1rem; border-radius: 6px; text-align: center; border: 1px solid var(--border); }
        .summary-val { display: block; font-size: 1.5rem; font-weight: 700; }
        .summary-label { font-size: 0.75rem; text-transform: uppercase; color: #64748b; letter-spacing: 0.05em; }
        table { width: 100%; border-collapse: collapse; margin-top: 1rem; }
        th { text-align: left; background: #f1f5f9; padding: 0.75rem; border-bottom: 2px solid var(--border); }
        td { padding: 0.75rem; border-bottom: 1px solid var(--border); font-size: 0.875rem; vertical-align: top; }
        .col-id { width: 100px; color: #64748b; font-family: ui-monospace, SFMono-Regular, monospace; font-size: 0.75rem; word-break: break-all; }
        .col-name { width: 250px; font-weight: 500; word-break: break-all; }
        .col-status { width: 80px; text-align: center; }
        .col-purpose { min-width: 250px; }
        .col-security { width: 200px; }
        .cat-header { background: #f8fafc; padding: 0.5rem 1rem; margin-top: 2rem; border-left: 4px solid var(--primary); font-weight: 600; }
        .failure-box { background: #0f172a; color: #f8fafc; padding: 1rem; border-radius: 4px; overflow-x: auto; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 0.75rem; margin-bottom: 1rem; }
        .security-mark { color: var(--warning); font-weight: 600; }
        .status-icon { font-size: 1.125rem; }
    </style>
</head>
<body>
    <div class="container">
        <h1>` + title + `</h1>
        <div class="meta">
            Generated at: ` + summary.GeneratedAt.Format("2006-01-02 15:04:05 MST") + ` | 
            Status: <span class="status-badge ` + statusClass + `">` + statusText + `</span>
        </div>

        <div class="summary-grid">
            <div class="summary-card"><span class="summary-val">` + fmt.Sprint(summary.Total) + `</span><span class="summary-label">Total</span></div>
            <div class="summary-card"><span class="summary-val" style="color: var(--success)">` + fmt.Sprint(summary.Passed) + `</span><span class="summary-label">Passed</span></div>
            <div class="summary-card"><span class="summary-val" style="color: var(--danger)">` + fmt.Sprint(summary.Failed) + `</span><span class="summary-label">Failed</span></div>
            <div class="summary-card"><span class="summary-val">` + fmt.Sprint(summary.Skipped) + `</span><span class="summary-label">Skipped</span></div>
            <div class="summary-card"><span class="summary-val">` + fmt.Sprintf("%.1f%%", rate) + `</span><span class="summary-label">Pass Rate</span></div>
        </div>

        <h2>Test Results</h2>`)

	categories := make(map[string][]FinalTestResult)
	for _, r := range summary.Results {
		cat := r.Annotations.Category
		if cat == "" {
			cat = "Uncategorized"
		}
		categories[cat] = append(categories[cat], r)
	}

	order := []string{"AuthN", "AuthZ", "Tenant", "OAuth2", "OIDC", "Audit", "Auth API", "Tenant API", "API", "SYSTEM Tests", "E2E Tests", "Other", "Uncategorized"}
	for _, cat := range order {
		tests, ok := categories[cat]
		if !ok || len(tests) == 0 {
			continue
		}

		sb.WriteString(`<div class="cat-header">` + cat + `</div>
        <table>
            <thead>
                <tr>
                    <th class="col-id">ID</th>
                    <th class="col-name">Test Name</th>
                    <th class="col-status">Status</th>
                    <th class="col-purpose">Purpose</th>
                    <th class="col-security">Security</th>
                </tr>
            </thead>
            <tbody>`)
		for _, t := range tests {
			icon := "✅"
			if t.Status == "fail" {
				icon = "❌"
			} else if t.Status == "skip" {
				icon = "⏭️"
			} else if t.Status == "not run" {
				icon = "⚪"
			}

			security := t.Annotations.Security
			if security != "" {
				security = `<span class="security-mark">🛡️ ` + security + `</span>`
			}

			sb.WriteString(`<tr>
                    <td class="col-id">` + t.Annotations.TestCaseID + `</td>
                    <td class="col-name"><code>` + t.Name + `</code></td>
                    <td class="col-status"><span class="status-icon">` + icon + `</span></td>
                    <td class="col-purpose">` + t.Annotations.Purpose + `</td>
                    <td class="col-security">` + security + `</td>
                </tr>`)
		}
		sb.WriteString(`</tbody></table>`)
	}

	if summary.Failed > 0 {
		sb.WriteString(`<h2>Failure Details</h2>`)
		for _, t := range summary.Results {
			if t.Status == "fail" {
				sb.WriteString(`<h3>` + t.Name + `</h3>
                <div class="failure-box"><pre>` + t.Failure + `</pre></div>`)
			}
		}
	}

	sb.WriteString(`
        <p style="margin-top: 3rem; color: #64748b; font-size: 0.75rem; text-align: center;">
            &copy; ` + fmt.Sprint(time.Now().Year()) + ` OpenTrusty Project | Generated by Test Infrastructure
        </p>
    </div>
</body>
</html>`)

	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(sb.String()), 0644)
}
