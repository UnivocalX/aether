#!/usr/bin/env bash
set -euo pipefail

# --- CONFIGURATION ---
# Default pre-release identifier if none provided for prerelease mode
DEFAULT_PRERELEASE_ID="rc"

# --- FUNCTIONS ---
usage() {
  echo "Usage: $0 [major|minor|patch|prerelease] [preid (optional, default=$DEFAULT_PRERELEASE_ID)]"
  exit 1
}

# Check dependencies
command -v git >/dev/null 2>&1 || { echo "git not found"; exit 1; }

# Get latest tag (strip 'v' prefix if exists)
get_latest_version() {
  git fetch --tags >/dev/null 2>&1 || true
  latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
  echo "${latest_tag#v}"
}

# Increment semantic version
bump_version() {
  local version="$1"
  local part="$2"
  local preid="$3"

  # Split version into components
  IFS='.-' read -r major minor patch pre <<<"$version"

  major=${major:-0}
  minor=${minor:-0}
  patch=${patch:-0}

  case "$part" in
    major)
      ((major++))
      minor=0
      patch=0
      new_version="${major}.${minor}.${patch}"
      ;;
    minor)
      ((minor++))
      patch=0
      new_version="${major}.${minor}.${patch}"
      ;;
    patch)
      ((patch++))
      new_version="${major}.${minor}.${patch}"
      ;;
    prerelease)
      if [[ -z "$pre" ]]; then
        new_version="${major}.${minor}.${patch}-${preid}.1"
      else
        # Extract preid and increment number
        prefix="${pre%%[0-9]*}"
        num="${pre##*[!0-9]}"
        ((num++))
        new_version="${major}.${minor}.${patch}-${prefix}${num}"
      fi
      ;;
    *)
      usage
      ;;
  esac

  echo "$new_version"
}

# --- MAIN ---
[[ $# -lt 1 ]] && usage

bump_type="$1"
preid="${2:-$DEFAULT_PRERELEASE_ID}"

current_version=$(get_latest_version)
new_version=$(bump_version "$current_version" "$bump_type" "$preid")

echo "Current version: v$current_version"
echo "New version:     v$new_version"

# Confirm and tag
read -p "Create and push tag v$new_version? [y/N]: " confirm
if [[ "$confirm" =~ ^[Yy]$ ]]; then
  git tag -a "v$new_version" -m "chore(release): v$new_version"
  git push origin "v$new_version"
  echo "✅ Tag v$new_version pushed."
else
  echo "❌ Aborted."
fi