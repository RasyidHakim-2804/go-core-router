# Documentation website

`docs/` contains a buildless static website for GitHub Pages. All assets are local.

After changing the root README, regenerate and commit the HTML:

```sh
python3 scripts/build_docs.py
```

The generator supports the Markdown constructs currently used by README (headings,
paragraphs, lists, fenced code, inline code, bold and links), not arbitrary Markdown.

Preview locally:

```sh
python3 -m http.server 8765 --directory docs
```

Open http://localhost:8765. To publish, push these files and select Settings → Pages
→ Deploy from a branch → main → /docs → Save in GitHub. No npm install or Go server
is required. `docs/.nojekyll` ensures the static files are served as generated.
