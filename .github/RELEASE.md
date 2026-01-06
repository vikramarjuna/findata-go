# Release Process

This document describes how to create a new release of findata-go.

## Prerequisites

- All tests must pass: `make test`
- All linting must pass: `make lint`
- CHANGELOG.md must be updated
- All changes must be committed and pushed

## Semantic Versioning

We follow [Semantic Versioning](https://semver.org/):

- **MAJOR** (v1.0.0 → v2.0.0): Breaking changes
- **MINOR** (v1.0.0 → v1.1.0): New features (backward compatible)
- **PATCH** (v1.0.0 → v1.0.1): Bug fixes (backward compatible)

### Pre-release versions

- **Alpha**: v1.0.0-alpha.1 (early testing)
- **Beta**: v1.0.0-beta.1 (feature complete, testing)
- **RC**: v1.0.0-rc.1 (release candidate)

## Release Steps

### 1. Update CHANGELOG.md

Move items from `[Unreleased]` to a new version section:

```markdown
## [1.0.0] - 2026-01-06

### Added
- New feature X
- New feature Y

### Fixed
- Bug fix Z

## [Unreleased]

(empty for now)
```

### 2. Update VERSION file

```bash
echo "v1.0.0" > VERSION
```

### 3. Commit changes

```bash
git add CHANGELOG.md VERSION
git commit -m "chore: prepare release v1.0.0"
git push origin main
```

### 4. Create and push tag

```bash
# Create tag locally
make tag-version VERSION=v1.0.0

# Push tag to GitHub
git push origin v1.0.0
```

### 5. Automated release

GitHub Actions will automatically:
- Run all tests
- Run linter
- Create a GitHub release
- Generate release notes
- Publish the release

### 6. Verify release

1. Go to https://github.com/Vikramarjuna/findata-go/releases
2. Verify the release was created
3. Check that release notes are correct
4. Test installation: `go get github.com/Vikramarjuna/findata-go@v1.0.0`

## Quick Commands

```bash
# Check current version
make version

# Run all checks before release
make test
make lint

# Create a patch release (v1.0.0 → v1.0.1)
make tag-version VERSION=v1.0.1
git push origin v1.0.1

# Create a minor release (v1.0.0 → v1.1.0)
make tag-version VERSION=v1.1.0
git push origin v1.1.0

# Create a major release (v1.0.0 → v2.0.0)
make tag-version VERSION=v2.0.0
git push origin v2.0.0

# Create a pre-release
make tag-version VERSION=v1.0.0-beta.1
git push origin v1.0.0-beta.1
```

## Rollback

If you need to delete a tag:

```bash
# Delete local tag
git tag -d v1.0.0

# Delete remote tag
git push origin :refs/tags/v1.0.0

# Delete GitHub release manually from the web UI
```

## Troubleshooting

### Tag already exists

```bash
# Delete and recreate
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0
make tag-version VERSION=v1.0.0
git push origin v1.0.0
```

### CI fails on release

1. Check the Actions tab on GitHub
2. Fix the issue
3. Delete the tag and release
4. Create a new tag with a patch version

