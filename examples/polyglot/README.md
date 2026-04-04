# Polyglot PI Example

Demonstrates mixing bash, Python, and TypeScript steps with pipe support (`pipe_to: next`).

## Automations

| Name | Description | Features Shown |
|------|-------------|----------------|
| `text/reverse` | Reverse a string | Inline Python step, argument passing |
| `text/transform` | Format text into a numbered box | `.py` file step (script beside automation), `pipe_to: next` |
| `data/generate` | Generate JSON data | Inline TypeScript step |
| `data/format` | Format a leaderboard from JSON | `.ts` file step, `run:` step piped to TypeScript |
| `pipeline/etl` | CSV → JSON → formatted output | Three-step pipe chain: bash → Python → TypeScript |
| `pipeline/wordcount` | Count words in text | Two-step pipe: bash → Python |

## Prerequisites

- Python 3 installed (`python3` in PATH)
- [tsx](https://github.com/privatenumber/tsx) installed (`npm install -g tsx`)

## Usage

```bash
# List all automations
pi list

# Inline Python — reverse text
pi run text/reverse
pi run text/reverse "automation"

# Python file — box formatter with piped input
pi run text/transform

# Inline TypeScript — generate JSON
pi run data/generate

# TypeScript file — leaderboard from piped JSON
pi run data/format

# Three-step ETL pipeline (bash → Python → TypeScript)
pi run pipeline/etl

# Two-step word count (bash → Python)
pi run pipeline/wordcount
```
