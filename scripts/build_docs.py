"""Render the README's supported Markdown subset into dependency-free Pages docs.
Run from any directory: python3 scripts/build_docs.py
"""
from pathlib import Path
import html
import re

ROOT = Path(__file__).resolve().parents[1]
REPO = 'https://github.com/RasyidHakim-2804/go-core-router'

def inline(value):
    tokens = []
    def protect(match):
        tokens.append('<code>' + html.escape(match[1]) + '</code>')
        return f'CODETOKEN{len(tokens)-1}END'
    value = re.sub(r'`([^`]+)`', protect, value)
    value = html.escape(value)
    value = re.sub(r'\*\*(.+?)\*\*', r'<strong>\1</strong>', value)
    value = re.sub(r'\[([^]]+)\]\(([^)]+)\)', lambda m: '<a href="' + (m[2] if m[2].startswith(('https://', '#')) else REPO + '/blob/main/' + m[2]) + '">' + m[1] + '</a>', value)
    for i, token in enumerate(tokens):
        value = value.replace(f'CODETOKEN{i}END', token)
    return value

lines = (ROOT / 'README.md').read_text().splitlines()
body, nav, paragraph, code = [], [], [], None
in_list = False

def flush():
    if paragraph:
        body.append('<p>' + inline(' '.join(paragraph)) + '</p>')
        paragraph.clear()

for line in lines:
    if code is not None:
        if line.startswith('```'):
            body.append('<div class="code-block"><div class="code-toolbar"><span>' + html.escape(language or 'text') + '</span><button type="button" class="copy" aria-label="Copy code">Copy</button></div><pre tabindex="0"><code>' + html.escape('\n'.join(code)) + '</code></pre></div>')
            code = None
        else:
            code.append(line)
        continue
    if in_list and not line.startswith('- '):
        body.append('</ul>'); in_list = False
    if line.startswith('```'):
        flush(); code = []; language = line[3:]; continue
    match = re.match(r'^(#{1,3}) (.+)$', line)
    if match:
        flush()
        level, title = len(match[1]), match[2]
        slug = re.sub(r'[^a-z0-9]+', '-', title.lower()).strip('-')
        if level == 1:
            body.append('<div class="eyebrow">DOCUMENTATION / OVERVIEW</div><h1>Go Core Router</h1>')
        else:
            short = {'Chaining without modifying the original router': 'Immutable chaining', 'Serving and configuration lifecycle': 'Serving & lifecycle', 'Running the complete example': 'Example application'}.get(title, title)
            nav.append(f'<a class="level-{level}" href="#{slug}">{html.escape(short)}</a>')
            body.append(f'<h{level} id="{slug}"><a class="heading-link" href="#{slug}" aria-label="Link to {html.escape(title)}">#</a>{html.escape(title)}</h{level}>')
    elif line.startswith('- '):
        flush()
        if not in_list:
            body.append('<ul>'); in_list = True
        body.append('<li>' + inline(line[2:]) + '</li>')
    elif not line.strip():
        flush()
    else:
        paragraph.append(line)
flush()
page = '''<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Go Core Router — Documentation</title><meta name="description" content="A lightweight Go router. Learn routes, middleware, immutable chaining, groups, prefixes and wildcards.">
<script src="./theme.js"></script><link rel="stylesheet" href="./style.css"><script src="./script.js" defer></script></head>
<body><a class="skip" href="#content">Skip to content</a>
<header class="mobile-header"><a href="#content">Go Core Router<span> / docs</span></a><button id="menu" aria-expanded="false" aria-controls="sidebar">Menu</button></header>
<aside id="sidebar"><a class="brand" href="#content"><span class="brand-icon">↳</span><span>Go Core Router<small>DOCUMENTATION</small></span></a>
<div class="version"><span class="dot"></span> Go 1.23.1+ <span>MIT licensed</span></div>
<label class="search-label" for="search">Find a topic</label><input id="search" type="search" placeholder="Filter topics…" autocomplete="off">
<nav aria-label="Documentation">NAV</nav><p id="empty" hidden>No matching topics.</p>
<a class="repository" href="REPO">View on GitHub <span>↗</span></a></aside>
<main id="content" tabindex="-1"><div class="topline"><span>THE GO ROUTING LIBRARY</span><div class="top-actions"><button id="theme-toggle" class="theme-toggle" type="button" aria-label="Dark mode" aria-pressed="false">Dark mode</button><a href="REPO">Source code ↗</a></div></div><article>BODY</article>
<footer>Go Core Router · <a href="REPO/blob/main/LICENSE">MIT License</a><span>Built on net/http.</span></footer></main>
<div id="announce" class="sr-only" role="status" aria-live="polite"></div></body></html>'''
page = page.replace('NAV', '\n'.join(nav)).replace('BODY', '\n'.join(body)).replace('REPO', REPO)
(ROOT / 'docs' / 'index.html').write_text(page)
