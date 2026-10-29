#!/bin/bash
# Repair corrupted node_modules packages by re-extracting registry tarballs.
# Usage: bash scripts/repair-package.sh <pkg-name> [<pkg-name> ...]
set -euo pipefail
cd "$(dirname "$0")/.."

for pkg in "$@"; do
  lock_entry=$(grep -A 4 "\"node_modules/$pkg\"" package-lock.json | head -8)
  version=$(printf '%s' "$lock_entry" | grep -o '"version": "[^"]*"' | head -1 | cut -d'"' -f4)
  resolved=$(printf '%s' "$lock_entry" | grep -o '"resolved": "[^"]*"' | head -1 | cut -d'"' -f4)
  if [ -z "$version" ] || [ -z "$resolved" ]; then
    echo "SKIP $pkg (not found in package-lock.json)"
    continue
  fi
  echo "Repairing $pkg@$version from $resolved"
  target="node_modules/$pkg"
  mkdir -p "$target"
  curl -fsSL "$resolved" -o /tmp/repair-pkg.tgz
  rm -rf /tmp/repair-pkg-dir
  mkdir -p /tmp/repair-pkg-dir
  tar -xzf /tmp/repair-pkg.tgz -C /tmp/repair-pkg-dir
  cp -R /tmp/repair-pkg-dir/package/. "$target/"
  echo "OK $pkg ($(find "$target" -type f | wc -l | tr -d ' ') files)"
done
