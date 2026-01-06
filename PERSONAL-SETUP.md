# Personal Development Setup

This file contains optional personal setup scripts that are **NOT** required for contributing to the project. These are tools you can use to enhance your personal development workflow.

> **Note:** These scripts are in `.gitignore` and will not be committed to the repository.

## Commit Message Spell Check

A personal git hook that spell checks your commit messages before committing.

### Setup

1. **Install a spell checker** (if not already installed):
   ```bash
   # Option 1: aspell (simple, lightweight)
   brew install aspell
   
   # Option 2: cspell (advanced, configurable)
   npm install -g cspell
   ```

2. **Run the setup script**:
   ```bash
   bash setup-personal-spell-check.sh
   ```

### Usage

Once installed, the hook runs automatically on every commit:

```bash
# This will fail if there are spelling errors
git commit -m "fix: corect the typo"
# ❌ Spelling errors found:
#   • corect

# This will succeed
git commit -m "fix: correct the typo"
# ✅ Commit successful
```

### Bypass the Hook

If you need to commit without spell checking:

```bash
git commit --no-verify -m "your message"
```

### Uninstall

To remove the hook:

```bash
rm .git/hooks/commit-msg
```

## Why This is Personal

Git hooks are **local** and **personal** because:

- ✅ Different developers have different workflows
- ✅ Not everyone has the same tools installed
- ✅ Hooks can slow down commits for some people
- ✅ Personal preference - some people prefer to spell check manually

**Best Practice:** Use CI/CD for checks that **must** pass for everyone. Use local hooks for personal productivity enhancements.

## Other Personal Tools

You can create other personal setup scripts following the same pattern:

1. Create a script named `setup-personal-*.sh`
2. It will be automatically ignored by git (see `.gitignore`)
3. Document it in this file for your future reference

