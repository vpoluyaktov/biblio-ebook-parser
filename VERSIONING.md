# Versioning Guide

## Current Version

**v0.1.0** - Initial release with EPUB and FB2 parsers

## Semantic Versioning

This library follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version (v1.0.0, v2.0.0): Breaking API changes
- **MINOR** version (v0.1.0, v0.2.0): New features, backward compatible
- **PATCH** version (v0.1.1, v0.1.2): Bug fixes, backward compatible

## Creating a New Release

### 1. Make your changes and commit them

```bash
git add .
git commit -m "fix: your bug fix description"
git push origin main
```

### 2. Tag the release

For a **patch** (bug fix):
```bash
git tag -a v0.1.1 -m "Bug fix: description"
git push origin v0.1.1
```

For a **minor** (new feature):
```bash
git tag -a v0.2.0 -m "Feature: description"
git push origin v0.2.0
```

For a **major** (breaking change):
```bash
git tag -a v1.0.0 -m "Breaking: description"
git push origin v1.0.0
```

### 3. Update consuming repositories

In `biblio-ebooks-catalog` and `biblio-audiobook-builder-tts`:

```bash
go get github.com/vpoluyaktov/biblio-ebook-parser@v0.1.1
go mod tidy
git add go.mod go.sum
git commit -m "chore: update parser to v0.1.1"
```

### 4. Rebuild and deploy

```bash
cd /home/ubuntu/git/biblio/biblio-hub/scripts
./stop_stack.sh
./rebuild_stack.sh
./start_stack.sh
```

## Version History

### v0.1.0 (2026-02-17)

Initial release:
- EPUB parser with TOC and spine-based chapter extraction
- FB2 parser with XML sanitization and charset handling
- Element-based content model (Paragraph, Heading, Image, etc.)
- HTML renderer for web readers
- PlainText renderer for TTS
- Fixed regex backreference error in heading extraction
- Fixed FB2 subsection handling

## Checking Current Version

In consuming repositories, check `go.mod`:
```bash
grep biblio-ebook-parser go.mod
```

Should show:
```
github.com/vpoluyaktov/biblio-ebook-parser v0.1.0
```
