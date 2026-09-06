#!/usr/bin/env bash
# Add NestedBOM under SubBOM and verify Top -> Sub -> Nested graph.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DT_BASE_URL="${DT_BASE_URL:-http://localhost:8082}"
DT_API_KEY="${DT_API_KEY:?Set DT_API_KEY}"
RC_BASE_URL="${RC_BASE_URL:-http://localhost:8080}"

TOP_UUID="${TOP_UUID:-673f7d7f-ec05-4289-a3f6-f014e87e3426}"
SUB_UUID="${SUB_UUID:-b118327a-dc0e-42ff-9e43-2bb7603bf2b9}"
BRANCH_ID="${BRANCH_ID:-44b77ad5-397a-413d-a471-cadb4e297577}"

NESTED_NAME="${NESTED_NAME:-NestedBOM}"
NESTED_PROJECT_VERSION="${NESTED_PROJECT_VERSION:-0.9.0}"
NESTED_BOM_VERSION="${NESTED_BOM_VERSION:-2}"
SUB_PROJECT_VERSION="${SUB_PROJECT_VERSION:-2.1.0}"
SUB_BOM_VERSION="${SUB_BOM_VERSION:-6}"

dt() {
  local method="$1"; shift
  local path="$1"; shift
  curl -sS -X "$method" \
    -H "X-Api-Key: ${DT_API_KEY}" \
    -H "Content-Type: application/json" \
    -H "Accept: application/json" \
    "${DT_BASE_URL}${path}" \
    "$@"
}

lookup_project() {
  local name="$1" version="$2"
  dt GET "/api/v1/project?name=$(python3 -c "import urllib.parse; print(urllib.parse.quote('''$name'''))")" \
    | python3 -c "
import sys, json
data = json.load(sys.stdin)
items = data if isinstance(data, list) else []
for p in items:
    if p.get('name') == '''$name''' and (p.get('version') or '') == '''$version''':
        print(p['uuid']); raise SystemExit(0)
raise SystemExit('project not found: %s %s' % ('''$name''', '''$version'''))
"
}

echo "==> Ensuring NestedBOM project exists"
NESTED_JSON="$(dt PUT /api/v1/project -d "{\"name\":\"${NESTED_NAME}\",\"version\":\"${NESTED_PROJECT_VERSION}\",\"classifier\":\"APPLICATION\",\"active\":true}" || true)"
NESTED_UUID="$(echo "$NESTED_JSON" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('uuid',''))" 2>/dev/null || true)"
if [[ -z "$NESTED_UUID" ]]; then
  NESTED_UUID="$(lookup_project "$NESTED_NAME" "$NESTED_PROJECT_VERSION")"
fi
echo "    NestedBOM uuid=$NESTED_UUID"

NESTED_BOM_FILE="${ROOT_DIR}/testdata/boms/nested-bom.json"
SUB_BOM_FILE="${ROOT_DIR}/testdata/boms/sub-bom.json"

python3 - "$NESTED_BOM_FILE" "$SUB_BOM_FILE" "$NESTED_UUID" "$SUB_UUID" "$NESTED_BOM_VERSION" "$SUB_BOM_VERSION" "$NESTED_PROJECT_VERSION" "$SUB_PROJECT_VERSION" <<'PY'
import json, sys
nested_path, sub_path, nested_uuid, sub_uuid, nested_bom_ver, sub_bom_ver, nested_proj_ver, sub_proj_ver = sys.argv[1:]
nested = {
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "serialNumber": f"urn:uuid:{nested_uuid}",
  "version": int(nested_bom_ver),
  "metadata": {
    "component": {
      "type": "application",
      "name": "NestedBOM",
      "version": nested_proj_ver,
      "bom-ref": "nested-bom",
    }
  },
  "components": [
    {
      "type": "library",
      "name": "nested-lib",
      "version": "0.1.0",
      "bom-ref": "pkg:example/nested-lib@0.1.0",
    }
  ],
}
sub = {
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "serialNumber": f"urn:uuid:{sub_uuid}",
  "version": int(sub_bom_ver),
  "metadata": {
    "component": {
      "type": "application",
      "name": "SubBOM",
      "version": sub_proj_ver,
      "bom-ref": "sub-bom",
    }
  },
  "components": [
    {
      "type": "application",
      "name": "NestedBOM",
      "version": nested_proj_ver,
      "bom-ref": "nested-bom-ref",
      "externalReferences": [
        {"type": "bom", "url": f"urn:cdx:{nested_uuid}/{nested_bom_ver}"}
      ],
    },
    {
      "type": "library",
      "name": "example-lib",
      "version": "1.2.3",
      "bom-ref": "pkg:example/example-lib@1.2.3",
    },
  ],
}
with open(nested_path, "w", encoding="utf-8") as f:
    json.dump(nested, f, indent=2); f.write("\n")
with open(sub_path, "w", encoding="utf-8") as f:
    json.dump(sub, f, indent=2); f.write("\n")
PY

upload_bom() {
  local project_uuid="$1"
  local bom_file="$2"
  local b64
  b64="$(base64 -w0 "$bom_file")"
  dt PUT /api/v1/bom -d "{\"project\":\"${project_uuid}\",\"bom\":\"${b64}\"}"
  echo
}

echo "==> Uploading NestedBOM"
upload_bom "$NESTED_UUID" "$NESTED_BOM_FILE"
echo "==> Uploading SubBOM (link -> NestedBOM)"
upload_bom "$SUB_UUID" "$SUB_BOM_FILE"

echo "==> Waiting for DT processing"
sleep 4

echo "==> Verifying SubBOM export has NestedBOM link"
curl -sS -H "X-Api-Key: ${DT_API_KEY}" -H "Accept: application/vnd.cyclonedx+json" \
  "${DT_BASE_URL}/api/v1/bom/cyclonedx/project/${SUB_UUID}?format=json" \
  | python3 -c "
import sys, json
bom = json.load(sys.stdin)
print('  SubBOM version=', bom.get('version'))
for c in bom.get('components') or []:
    for r in c.get('externalReferences') or []:
        if r.get('type') == 'bom':
            print('  BOM-Link:', c.get('name'), '->', r.get('url'))
"

echo "==> Fetching Release Catalog graph"
GRAPH="$(curl -sS "${RC_BASE_URL}/api/v1/branches/${BRANCH_ID}/current/graph")"
echo "$GRAPH" | python3 -c "
import sys, json
g = json.load(sys.stdin)
md = g.get('metadata') or {}
print('  nodes=', md.get('totalNodes'), 'edges=', md.get('totalEdges'), 'unresolved=', md.get('unresolvedLinks'))
nodes = {n['id']: n for n in g.get('nodes') or []}
for n in g.get('nodes') or []:
    print(f\"  - {n.get('projectName')} projectVersion={n.get('projectVersion')} bomVersion={n.get('bomVersion')} status={n.get('resolutionStatus')}\")
for e in g.get('edges') or []:
    src = nodes.get(e['sourceNodeId'], {})
    dst = nodes.get(e['targetNodeId'], {})
    print(f\"  edge: {src.get('projectName')} -> {dst.get('projectName')}\")
"

echo
echo "Done. UI: http://localhost:3000/branches/${BRANCH_ID}"
echo "NestedBOM uuid=${NESTED_UUID}"
