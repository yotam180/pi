"""Text transformation script: uppercase, add line numbers, wrap in a box."""
import sys

lines = []
for line in sys.stdin:
    lines.append(line.rstrip("\n"))

if not lines:
    lines = ["no input"]

width = max(len(l) for l in lines)
border = "+" + "-" * (width + 6) + "+"

print(border)
for i, line in enumerate(lines, 1):
    print(f"| {i:2d}. {line:<{width}} |")
print(border)
