#!/bin/bash
set -e

PORT="${1:-4444}"
BASE="http://localhost:${PORT}"

seed() {
  local id="$1"
  local body="$2"
  status=$(curl -s -o /dev/null -w "%{http_code}" -X POST "${BASE}/api/proposals" -H 'Content-Type: application/json' -d "$body")
  echo "  ${id}: ${status}"
}

echo "seeding proposals to ${BASE}..."

seed "abc-001" '{
  "id": "abc-001",
  "title": "Extract payment validation into its own module",
  "summary": "The checkout handler mixes HTTP concerns with payment validation logic. Validation rules are duplicated between the API and the webhook handler, and they have already diverged in subtle ways.",
  "plan": "## Problem\n\nThe checkout handler in checkout.go is 400 lines long and mixes request parsing, validation, payment gateway calls, and response formatting.\n\nThe same validation logic (amount bounds, currency checks, idempotency key format) is duplicated in stripe.go but with slightly different bounds.\n\n## Proposed Change\n\nExtract a payments/validate package that owns all validation rules.\n\n## Risks\n\nNeed to decide which bound is correct — requires product input.",
  "session_id": "fake-session-001"
}'

seed "abc-002" '{
  "id": "abc-002",
  "title": "Replace hand-rolled SQL query builder with parameterized queries",
  "summary": "The query builder uses string concatenation to build SQL queries. While inputs are escaped, the pattern is fragile and has already caused one bug where a column name with a hyphen broke the query.",
  "plan": "## Problem\n\npkg/db/query.go contains a buildQuery function that concatenates SQL strings with fmt.Sprintf.\n\n## Proposed Change\n\nReplace with standard parameterized queries using ? placeholders.\n\n## Risks\n\nLow. This is a well-understood pattern.",
  "session_id": "fake-session-002"
}'

seed "abc-003" '{
  "id": "abc-003",
  "title": "Consolidate three nearly-identical config loading functions",
  "summary": "There are three functions that load config from environment, file, and defaults. They share 80% of their logic but have drifted apart. One silently ignores parse errors while the others fail loudly.",
  "plan": "## Problem\n\nLoadFromEnv, LoadFromFile, and LoadDefaults were clearly copy-pasted from one original. LoadFromEnv swallows strconv.ParseInt errors while the others return the error.\n\n## Proposed Change\n\nUnify into a single Load(sources ...Source) function that applies sources in order.\n\n## Risks\n\nThe silent-swallow behavior might be intentional. Need to check with the team.",
  "session_id": "fake-session-003"
}'

echo "done"
