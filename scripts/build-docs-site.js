const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Configuration
const CONFIG = require('../docs/docs-config');
const REPO_ROOT = path.resolve(__dirname, '..');
const OUTPUT_DIR = path.resolve(REPO_ROOT, 'build_docs');
const VERSION = process.env.DOCS_VERSION || 'latest';
const ALL_VERSIONS = (process.env.ALL_VERSIONS || VERSION).split(',');

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
    <script src="https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"></script>
    <script>
        var API_MODE = "{{is_api}}" === "true";
        var basePath = API_MODE ? ".." : "."; 
    </script>
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
        .sidebar {
            width: 300px;
            background: var(--sidebar-bg);
            backdrop-filter: blur(12px);
            border-right: 1px solid var(--border);
            padding: 2rem;
            position: fixed;
            height: 100vh;
            overflow-y: auto;
            z-index: 1000;
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
            min-height: 100vh;
        }
        .markdown-body { line-height: 1.7; }
        .markdown-body h1 { font-family: 'Montserrat', sans-serif; margin-bottom: 2rem; color: var(--primary); }
        .markdown-body pre { background: #1e293b; padding: 1rem; border-radius: 8px; border: 1px solid var(--border); overflow-x: auto; }
        
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
        <a href="/" class="logo" onclick="event.preventDefault(); window.location.href=basePath + '/index.html'">
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg">
                <circle cx="16" cy="16" r="14" stroke="currentColor" stroke-width="3"/>
                <path d="M16 8V24M8 16H24" stroke="currentColor" stroke-width="3" stroke-linecap="round"/>
            </svg>
            OpenTrusty
        </a>
        <nav>{{nav}}</nav>
        <div class="version-selector">
            <div class="nav-title">Version</div>
            <select id="v-select" onchange="window.location.href=this.value">{{versions}}</select>
        </div>
    </div>
    <main class="content-wrapper">
        <div id="content" class="markdown-body">{{content}}</div>
    </main>
    <script id="md-content" type="text/template">{{raw_content}}</script>
    <script>
        var raw = document.getElementById('md-content').textContent;
        if (raw && raw.trim().length > 0) {
            document.getElementById('content').innerHTML = marked.parse(raw);
        }
        if (API_MODE) {
            var specPath = basePath + '/openapi.json';
            var contentRoot = document.getElementById('content');
            // Use default ReDoc styling - no custom theme
            Redoc.init(specPath, {}, contentRoot);
            
            var wrapper = document.querySelector('.content-wrapper');
            wrapper.style.padding = '0';
            wrapper.style.maxWidth = 'none';
            wrapper.style.backgroundColor = '#ffffff';
        }
    </script>
</body>
</html>
`;

function generateNav(currentFile) {
    var isApi = currentFile.indexOf('api/index.html') !== -1;
    var prefix = isApi ? '../' : '';

    return CONFIG.sections.map(function (section) {
        var items = section.items.map(function (item) {
            var isActive = item.file === currentFile;
            var href;
            if (item.file.indexOf('.md') !== -1) {
                href = prefix + item.file.replace(/\.md$/, '.html').replace(/\//g, '_');
            } else if (item.file.indexOf('api/index.html') !== -1) {
                href = prefix + 'api/index.html';
            } else {
                href = prefix + item.file.replace('docs/', '');
            }
            return '<a href="' + href + '" class="nav-link ' + (isActive ? 'active' : '') + '">' + item.title + '</a>';
        }).join('');

        return '<div class="nav-section"><div class="nav-title">' + section.title + '</div>' + items + '</div>';
    }).join('');
}

function build() {
    if (!fs.existsSync(OUTPUT_DIR)) fs.mkdirSync(OUTPUT_DIR, { recursive: true });

    var options = ALL_VERSIONS.map(function (v) {
        var selected = v === VERSION ? 'selected' : '';
        var vPath = v === 'latest' ? '/' : '/versions/' + v + '/';
        return '<option value="' + vPath + '" ' + selected + '>' + v + '</option>';
    }).join('');

    CONFIG.sections.forEach(function (section) {
        section.items.forEach(function (item) {
            var h = HTML_TEMPLATE.replace(/{{nav}}/g, generateNav(item.file)).replace(/{{versions}}/g, options);

            if (item.file.indexOf('.md') !== -1) {
                var fullPath = path.join(REPO_ROOT, item.file);
                if (!fs.existsSync(fullPath)) {
                    console.log('Warning: File not found: ' + fullPath);
                    return;
                }

                var raw = fs.readFileSync(fullPath, 'utf8');
                var slug = item.file.replace('.md', '.html').replace(/\//g, '_');

                var h2 = h.replace(/{{title}}/g, item.title)
                    .replace(/{{content}}/g, '')
                    .replace(/{{raw_content}}/g, raw)
                    .replace(/{{is_api}}/g, 'false');

                fs.writeFileSync(path.join(OUTPUT_DIR, slug), h2);
                console.log('Generated: ' + slug);
            } else if (item.file.indexOf('api/index.html') !== -1) {
                var apiDir = path.join(OUTPUT_DIR, 'api');
                if (!fs.existsSync(apiDir)) fs.mkdirSync(apiDir, { recursive: true });

                var h2 = h.replace(/{{title}}/g, item.title)
                    .replace(/{{content}}/g, '')
                    .replace(/{{raw_content}}/g, '')
                    .replace(/{{is_api}}/g, 'true');

                fs.writeFileSync(path.join(apiDir, 'index.html'), h2);
                console.log('Generated: api/index.html');
            }
        });
    });

    var first = CONFIG.sections[0].items[0];
    var firstSlug = first.file.replace('.md', '.html').replace(/\//g, '_');
    fs.copyFileSync(path.join(OUTPUT_DIR, firstSlug), path.join(OUTPUT_DIR, 'index.html'));
}

build();
