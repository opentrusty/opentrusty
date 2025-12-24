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
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Montserrat:wght@600;700&display=swap" rel="stylesheet">
    <script src="https://cdn.jsdelivr.net/npm/marked/marked.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/redoc/bundles/redoc.standalone.js"></script>
    <script>
        var API_MODE = "{{is_api}}" === "true";
        var basePath = API_MODE ? ".." : "."; 
    </script>
    <style>
        :root {
            --primary: #059669; /* Emerald 600 - Reliable Tech */
            --primary-dark: #065F46; /* Logo Color - Steady */
            --bg: #f8fafc; /* Content BG - Slate 50 */
            --sidebar-bg: #041e1a; /* Very Deep Emerald - Tech/Steady */
            --text: #0f172a; /* Slate 900 */
            --text-muted: #64748b; /* Slate 500 */
            --sidebar-text: #f0fdf4; /* Emerald 50 */
            --sidebar-text-muted: #94a3b8;
            --border: rgba(6, 95, 70, 0.1);
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
            border-right: 1px solid rgba(255,255,255,0.05);
            padding: 2.5rem 2rem;
            position: fixed;
            height: 100vh;
            overflow-y: auto;
            z-index: 1000;
            display: flex;
            flex-direction: column;
            color: var(--sidebar-text);
        }
        .logo {
            font-family: 'Montserrat', sans-serif;
            font-size: 1.25rem;
            margin-bottom: 3rem;
            display: flex;
            align-items: center;
            gap: 0.75rem;
            color: #fff;
            text-decoration: none;
            font-weight: 700;
        }
        .logo svg { flex-shrink: 0; }
        
        .nav-section { margin-bottom: 2rem; }
        .nav-title {
            text-transform: uppercase;
            font-size: 0.7rem;
            letter-spacing: 0.1rem;
            color: var(--sidebar-text-muted);
            margin-bottom: 1rem;
            font-weight: 700;
        }
        .nav-link {
            display: block;
            padding: 0.6rem 0;
            color: var(--sidebar-text-muted);
            text-decoration: none;
            font-size: 0.9rem;
            transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
            border-left: 2px solid transparent;
            padding-left: 0;
        }
        .nav-link:hover { 
            color: #fff; 
            padding-left: 0.5rem;
        }
        .nav-link.active { 
            color: var(--primary); 
            font-weight: 600; 
            border-left: 2px solid var(--primary);
            padding-left: 0.75rem;
        }

        .content-wrapper {
            margin-left: 300px;
            flex: 1;
            padding: 5rem 6rem;
            max-width: 1200px;
            min-height: 100vh;
        }
        .markdown-body { 
            line-height: 1.8; 
            font-size: 1.05rem;
            color: #334155;
        }
        .markdown-body h1 { 
            font-family: 'Montserrat', sans-serif; 
            margin-bottom: 2.5rem; 
            color: var(--primary-dark);
            font-weight: 700;
            font-size: 2.5rem;
            letter-spacing: -0.02em;
        }
        .markdown-body h2 {
            border-bottom: 1px solid var(--border);
            padding-bottom: 0.5rem;
            margin-top: 3rem;
            color: var(--primary-dark);
            font-family: 'Montserrat', sans-serif;
            font-weight: 600;
        }
        .markdown-body p { margin-bottom: 1.5rem; }
        .markdown-body pre { 
            background: #0f172a; 
            padding: 1.5rem; 
            border-radius: 12px; 
            border: 1px solid var(--border); 
            overflow-x: auto;
            color: #e2e8f0;
            box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.05);
        }
        .markdown-body code {
            background: rgba(5, 150, 105, 0.05);
            color: var(--primary-dark);
            padding: 0.2rem 0.4rem;
            border-radius: 4px;
            font-size: 0.9em;
        }
        .markdown-body pre code {
            background: transparent;
            color: inherit;
            padding: 0;
        }
        
        .version-selector {
            margin-top: auto;
            padding-top: 2rem;
            border-top: 1px solid rgba(255,255,255,0.05);
        }
        select {
            background: rgba(255,255,255,0.05);
            color: #fff;
            border: 1px solid rgba(255,255,255,0.1);
            padding: 0.6rem;
            border-radius: 8px;
            width: 100%;
            cursor: pointer;
            font-size: 0.85rem;
            appearance: none;
            outline: none;
        }
        select:hover {
            background: rgba(255,255,255,0.08);
        }
    </style>
</head>
<body>
    <div class="sidebar">
        <a href="/" class="logo" onclick="event.preventDefault(); window.location.href=basePath + '/index.html'">
            <svg width="28" height="28" viewBox="0 0 90 90" xmlns="http://www.w3.org/2000/svg">
                <rect width="90" height="90" rx="20" fill="#065F46" />
                <path d="M15,25 H75 V40 H53 V70 H37 V40 H15 Z" fill="white" />
                <rect x="37" y="73" width="16" height="4" rx="1" fill="white" opacity="0.6" />
            </svg>
            OpenTrusty
        </a>
        <nav>{{nav}}</nav>
        <div class="version-selector">
            <div class="nav-title">Release Version</div>
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
            
            Redoc.init(specPath, {
                expandResponses: '200,201',
                requiredPropsFirst: true,
                showExtensions: true,
                scrollYOffset: 0,
                hideDownloadButton: false,
                theme: {
                    colors: {
                        primary: { main: '#059669' },
                        text: { primary: '#334155' }
                    },
                    typography: {
                        fontFamily: 'Inter, sans-serif',
                        headings: {
                            fontFamily: 'Montserrat, sans-serif'
                        }
                    },
                    rightPanel: {
                        backgroundColor: '#0f172a'
                    }
                }
            }, contentRoot);
            
            var wrapper = document.querySelector('.content-wrapper');
            wrapper.style.padding = '0';
            wrapper.style.maxWidth = 'none';
            wrapper.style.marginLeft = '300px';
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
                    console.log('❌ Error: Source file not found: ' + fullPath);
                    return;
                }
                console.log('📖 Processing: ' + item.file);

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
