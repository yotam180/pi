#!/usr/bin/env bash
set -euo pipefail

# Generates llms-ctx.txt by concatenating all .md files referenced in llms.txt.
# Run after `npm run build` so that dist/ contains the .md mirrors.

DIST_DIR="${1:-dist}"
LLMS_TXT="$DIST_DIR/llms.txt"
OUTPUT="$DIST_DIR/llms-ctx.txt"

if [ ! -f "$LLMS_TXT" ]; then
  echo "error: $LLMS_TXT not found. Run 'npm run build' first." >&2
  exit 1
fi

{
  echo "# PI — Full Documentation Context"
  echo ""
  echo "This file contains all PI documentation concatenated as clean Markdown."
  echo "Paste it into any LLM with a large context window for comprehensive PI knowledge."
  echo ""

  # Extract .md links from llms.txt (relative paths like getting-started/introduction.md)
  grep -oE '[a-z][-a-z0-9/]*\.md' "$LLMS_TXT" | while read -r md_path; do
    full_path="$DIST_DIR/$md_path"
    if [ -f "$full_path" ]; then
      echo "---"
      echo ""
      cat "$full_path"
      echo ""
      echo ""
    fi
  done
} > "$OUTPUT"

echo "Generated $OUTPUT ($(wc -c < "$OUTPUT" | tr -d ' ') bytes)"
