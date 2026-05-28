#!/usr/bin/env sh
set -eu

API="docs/api.yaml"
OUT="handlers/static"

mkdir -p "$OUT/api"

redocly build-docs "$API" \
  --output "$OUT/api/index.html" \
  >/dev/null 2>&1

redocly bundle "$API" \
  --remove-unused-components \
  --output "$OUT/openapi.pretty.json" \
  >/dev/null 2>&1

redocly bundle "$API" \
  --dereferenced \
  --remove-unused-components \
  --output "$OUT/openapi.deref.pretty.json" \
  >/dev/null 2>&1

node -e "
const fs = require('fs');

for (const [input, output] of [
  ['$OUT/openapi.pretty.json', '$OUT/openapi.json'],
  ['$OUT/openapi.deref.pretty.json', '$OUT/openapi.deref.json'],
]) {
  const json = JSON.parse(fs.readFileSync(input, 'utf8'));
  fs.writeFileSync(output, JSON.stringify(json));
}
"

rm "$OUT/openapi.pretty.json" "$OUT/openapi.deref.pretty.json"
