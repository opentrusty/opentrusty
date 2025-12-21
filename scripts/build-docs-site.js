const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Configuration
const CONFIG = require('../docs/docs-config');
const REPO_ROOT = path.resolve(__dirname, '..');
const OUTPUT_DIR = path.resolve(REPO_ROOT, 'build_docs');
const VERSION = process.env.DOCS_VERSION || 'latest';

// Ensure marked is available for markdown parsing
// We'll use a cdn-based approach in the template for simplicity OR expect it to be installed
// For a standalone script, we might need a minimal internal parser or a common dependency.
// Given common runner environments, let's assume we can npm install it if needed, 
// but for the most stable "premium" look, we'll use a template that renders markdown via a client library (marked.js)
// to keep the generator script dependency-free and lightweight.

const HTML_TEMPLATE = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{title}} - OpenTrusty Docs</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Montserrat:wght@700&display=swap" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <style>
        :root {
            --primary: #4A90E2;
            --bg: #0f172a;
            --sidebar-bg: rgba(30, 41, 59, 0.7);
            --text: #f8fafc;
            --text-muted: #94a3b8;
            --border: rgba(255, 255, 255, 0.1);
        }
        * { box-sizing: border-box; }
        body {
            font-family: 'Inter', sans-serif;
            background-color: var(--bg);
            color: var(--text);
            margin: 0;
            display: flex;
            min-height: 100vh;
        }
        /* Glassmorphism Sidebar */
        .sidebar {
            width: 300px;
            background: var(--sidebar-bg);
            backdrop-filter: blur(12px);
            border-right: 1px solid var(--border);
            padding: 2rem;
            position: fixed;
            height: 100vh;
            overflow-y: auto;
        }
        .logo {
            font-family: 'Montserrat', sans-serif;
            font-size: 1.5rem;
            margin-bottom: 2rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            color: var(--primary);
            text-decoration: none;
        }
        .nav-section { margin-bottom: 1.5rem; }
        .nav-title {
            text-transform: uppercase;
            font-size: 0.75rem;
            letter-spacing: 0.05em;
            color: var(--text-muted);
            margin-bottom: 0.75rem;
            font-weight: 600;
        }
        .nav-link {
            display: block;
            padding: 0.5rem 0;
            color: var(--text);
            text-decoration: none;
            font-size: 0.95rem;
            transition: color 0.2s;
        }
        .nav-link:hover { color: var(--primary); }
        .nav-link.active { color: var(--primary); font-weight: 600; }

        .content-wrapper {
            margin-left: 300px;
            flex: 1;
            padding: 4rem;
            max-width: 1000px;
        }
        .markdown-body { line-height: 1.7; }
        .markdown-body h1 { font-family: 'Montserrat', sans-serif; margin-bottom: 2rem; color: var(--primary); }
        .markdown-body pre { background: #1e293b; padding: 1rem; border-radius: 8px; overflow-x: auto; border: 1px solid var(--border); }
        .markdown-body code { font-family: 'ui-monospace', monospace; background: rgba(255,255,255,0.1); padding: 0.2rem 0.4rem; border-radius: 4px; }
        
        .version-selector {
            margin-top: auto;
            padding-top: 2rem;
            border-top: 1px solid var(--border);
        }
        select {
            background: #1e293b;
            color: var(--text);
            border: 1px solid var(--border);
            padding: 0.5rem;
            border-radius: 4px;
            width: 100%;
            cursor: pointer;
        }
    </style>
</head>
<body>
    <div class="sidebar">
        <a href="/" class="logo">
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="16" cy="16" r="14" stroke="currentColor" stroke-width="3"/>
                <path d="M16 8V24M8 16H24" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
            </svg>
            OpenTrusty
        </a>
        
        <nav>
            {{nav}}
        </nav>

        <div class="version-selector">
            <div class="nav-title">Version</div>
            <select id="v-select" onchange="window.location.href=this.value">
                {{versions}}
            </select>
        </div>
    </div>

    <main class="content-wrapper">
        <div id="content" class="markdown-body">
            {{content}}
        </div>
    </main>

    <script>
        // Simple markdown injection if content is raw
        const rawContent = \`{{raw_content}}\`;
        if (rawContent && rawContent.length > 0) {
            document.getElementById('content').innerHTML = marked.parse(rawContent);
        }
    </script>
</body>
</html>
`;

function generateNav(currentFile) {
    return CONFIG.sections.map(section => {
        const items = section.items.map(item => {
            const isActive = item.file === currentFile;
            // Handle cross-directory links correctly in the static output
            // For now, we'll assume a flat build structure or relative paths
            const slug = item.file.replace('.md', '.html').replace(/\//g, '_');
            return `<a href="${slug}" class="nav-link ${isActive ? 'active' : ''}">${item.title}</a>`;
        }).join('');

        return `
            <div class="nav-section">
                <div class="nav-title">${section.title}</div>
                ${items}
            </div>
        `;
    }).join('');
}

function build() {
    if (!fs.existsSync(OUTPUT_DIR)) fs.mkdirSync(OUTPUT_DIR, { recursive: true });

    // 1. Process all markdown files from config
    CONFIG.sections.forEach(section => {
        section.items.forEach(item => {
            if (item.file.endsWith('.md')) {
                const fullPath = path.join(REPO_ROOT, item.file);
                if (!fs.existsSync(fullPath)) {
                    console.warn(`Warning: File not found: ${fullPath}`);
                    return;
                }

                const content = fs.readFileSync(fullPath, 'utf8');
                const slug = item.file.replace('.md', '.html').replace(/\//g, '_');

                let html = HTML_TEMPLATE
                    .replace('{{title}}', item.title)
                    .replace('{{nav}}', generateNav(item.file))
                    .replace('{{content}}', '') // Client side marked will handle it if we pass raw
                    .replace('{{raw_content}}', content.replace(/`/g, '\\`').replace(/\$/g, '\\$'));

                // For now, placeholder for version list
                html = html.replace('{{versions}}', `<option value="#">${VERSION}</option>`);

                fs.writeFileSync(path.join(OUTPUT_DIR, slug), html);
                console.log(`Generated: ${slug}`);
            }
        });
    });

    // 2. Create index redirecting to the first architecture page
    const firstItem = CONFIG.sections[0].items[0];
    const indexSlug = firstItem.file.replace('.md', '.html').replace(/\//g, '_');
    fs.copyFileSync(path.join(OUTPUT_DIR, indexSlug), path.join(OUTPUT_DIR, 'index.html'));
}

build();
