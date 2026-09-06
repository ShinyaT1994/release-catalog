#!/usr/bin/env bash
# Seed TopBOM + SubBOM into Dependency-Track and wire Release Catalog Current State.
# Requires: curl, python3, base64
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
DT_BASE_URL="${DT_BASE_URL:-http://localhost:8082}"
DT_API_KEY="${DT_API_KEY:?Set DT_API_KEY (Dependency-Track API key)}"
RC_BASE_URL="${RC_BASE_URL:-http://localhost:8080}"

TOP_NAME="${TOP_NAME:-TopBOM}"
TOP_PROJECT_VERSION="${TOP_PROJECT_VERSION:-1.0.0}"
TOP_BOM_VERSION="${TOP_BOM_VERSION:-3}"
SUB_NAME="${SUB_NAME:-SubBOM}"
SUB_PROJECT_VERSION="${SUB_PROJECT_VERSION:-2.1.0}"
SUB_BOM_VERSION="${SUB_BOM_VERSION:-5}"

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

echo "==> Ensuring DT projects exist"
SUB_JSON="$(dt PUT /api/v1/project -d "{\"name\":\"${SUB_NAME}\",\"version\":\"${SUB_PROJECT_VERSION}\",\"classifier\":\"APPLICATION\",\"active\":true}")"
TOP_JSON="$(dt PUT /api/v1/project -d "{\"name\":\"${TOP_NAME}\",\"version\":\"${TOP_PROJECT_VERSION}\",\"classifier\":\"APPLICATION\",\"active\":true}")"

# If project already exists, DT returns 409 — look up by name+version
lookup_project() {
  local name="$1" version="$2"
  dt GET "/api/v1/project?name=$(python3 -c "import urllib.parse; print(urllib.parse.quote('''$name'''))")&version=$(python3 -c "import urllib.parse; print(urllib.parse.quote('''$version'''))")" \
    | python3 -c "
import sys, json
data = json.load(sys.stdin)
items = data if isinstance(data, list) else data.get('content', data.get('data', []))
for p in items:
    if p.get('name') == '''$name''' and (p.get('version') or '') == '''$version''':
        print(p['uuid']); raise SystemExit(0)
raise SystemExit('project not found: %s %s' % ('''$name''', '''$version'''))
"
}

extract_uuid() {
  local json="$1" name="$2" version="$3"
  if echo "$json" | python3 -c "import sys,json; json.load(sys.stdin)" >/dev/null 2>&1; then
    local uuid
    uuid="$(echo "$json" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('uuid',''))" 2>/dev/null || true)"
    if [[ -n "$uuid" ]]; then
      echo "$uuid"
      return
    fi
  fi
  lookup_project "$name" "$version"
}

SUB_UUID="$(extract_uuid "$SUB_JSON" "$SUB_NAME" "$SUB_PROJECT_VERSION")"
TOP_UUID="$(extract_uuid "$TOP_JSON" "$TOP_NAME" "$TOP_PROJECT_VERSION")"
echo "    SubBOM uuid=$SUB_UUID"
echo "    TopBOM uuid=$TOP_UUID"

SUB_BOM_FILE="${ROOT_DIR}/testdata/boms/sub-bom.json"
TOP_BOM_FILE="${ROOT_DIR}/testdata/boms/top-bom.json"

echo "==> Writing BOM fixtures with live UUIDs / versions"
python3 - "$SUB_BOM_FILE" "$TOP_BOM_FILE" "$SUB_UUID" "$TOP_UUID" "$SUB_BOM_VERSION" "$TOP_BOM_VERSION" "$SUB_PROJECT_VERSION" "$TOP_PROJECT_VERSION" <<'PY'
import json, sys
sub_path, top_path, sub_uuid, top_uuid, sub_bom_ver, top_bom_ver, sub_proj_ver, top_proj_ver = sys.argv[1:]
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
      "type": "library",
      "name": "example-lib",
      "version": "1.2.3",
      "bom-ref": "pkg:example/example-lib@1.2.3",
    }
  ],
}
top = {
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "serialNumber": f"urn:uuid:{top_uuid}",
  "version": int(top_bom_ver),
  "metadata": {
    "component": {
      "type": "application",
      "name": "TopBOM",
      "version": top_proj_ver,
      "bom-ref": "top-bom",
    }
  },
  "components": [
    {
      "type": "application",
      "name": "SubBOM",
      "version": sub_proj_ver,
      "bom-ref": "sub-bom-ref",
      "externalReferences": [
        {"type": "bom", "url": f"urn:cdx:{sub_uuid}/{sub_bom_ver}"}
      ],
    }
  ],
}
with open(sub_path, "w", encoding="utf-8") as f:
    json.dump(sub, f, indent=2)
    f.write("\n")
with open(top_path, "w", encoding="utf-8") as f:
    json.dump(top, f, indent=2)
    f.write("\n")
PY

upload_bom() {
  local project_uuid="$1"
  local bom_file="$2"
  local b64
  b64="$(base64 -w0 "$bom_file")"
  dt PUT /api/v1/bom -d "{\"project\":\"${project_uuid}\",\"bom\":\"${b64}\"}"
  echo
}

echo "==> Uploading SubBOM (CycloneDX version=${SUB_BOM_VERSION})"
upload_bom "$SUB_UUID" "$SUB_BOM_FILE"
echo "==> Uploading TopBOM (CycloneDX version=${TOP_BOM_VERSION}, link -> SubBOM/${SUB_BOM_VERSION})"
upload_bom "$TOP_UUID" "$TOP_BOM_FILE"

echo "==> Waiting for DT to process BOMs"
sleep 3

echo "==> Verifying exported BOMs (DT may rewrite serialNumber/version on import)"
TOP_SERIAL=""
TOP_EXPORTED_VERSION=""
for uuid in "$SUB_UUID" "$TOP_UUID"; do
  META="$(curl -sS -H "X-Api-Key: ${DT_API_KEY}" -H "Accept: application/vnd.cyclonedx+json, application/json" \
    "${DT_BASE_URL}/api/v1/bom/cyclonedx/project/${uuid}?format=json")"
  echo "$META" | python3 -c "
import sys, json
bom = json.load(sys.stdin)
print('  project', '''$uuid''', 'serial=', bom.get('serialNumber'), 'version=', bom.get('version'), 'components=', len(bom.get('components') or []))
for c in bom.get('components') or []:
    for r in c.get('externalReferences') or []:
        if r.get('type') == 'bom':
            print('    BOM-Link:', r.get('url'))
"
  if [[ "$uuid" == "$TOP_UUID" ]]; then
    TOP_SERIAL="$(echo "$META" | python3 -c "import sys,json; print(json.load(sys.stdin).get('serialNumber',''))")"
    TOP_EXPORTED_VERSION="$(echo "$META" | python3 -c "import sys,json; print(json.load(sys.stdin).get('version',1))")"
  fi
done

echo "==> Seeding Release Catalog product / current state"
PRODUCT="$(curl -sS -X POST -H "Content-Type: application/json" \
  "${RC_BASE_URL}/api/v1/products" \
  -d '{"name":"Top-Sub Demo","displayName":"TopBOM / SubBOM Demo","description":"Test product for TopBOM->SubBOM graph"}')"
PRODUCT_ID="$(echo "$PRODUCT" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")"
echo "    productId=$PRODUCT_ID"

BRANCHES="$(curl -sS "${RC_BASE_URL}/api/v1/products/${PRODUCT_ID}/branches")"
MAIN_ID="$(echo "$BRANCHES" | python3 -c "
import sys, json
branches = json.load(sys.stdin)
for b in branches:
    if b.get('type') == 'MAIN':
        print(b['id']); raise SystemExit(0)
print(branches[0]['id'])
")"
echo "    mainBranchId=$MAIN_ID"

# Prefer DT-exported serial/version for Current State display
CS_VERSION="${TOP_EXPORTED_VERSION:-$TOP_BOM_VERSION}"
CS_SERIAL="${TOP_SERIAL:-urn:uuid:${TOP_UUID}}"
CS="$(curl -sS -X PUT -H "Content-Type: application/json" \
  "${RC_BASE_URL}/api/v1/branches/${MAIN_ID}/current" \
  -d "{\"rootDtProjectUuid\":\"${TOP_UUID}\",\"rootBomSerialNumber\":\"${CS_SERIAL}\",\"rootBomVersion\":${CS_VERSION},\"sourceRevision\":\"testdata-top-sub\"}")"
echo "    currentState=$(echo "$CS" | python3 -c "import sys,json; d=json.load(sys.stdin); print('root=', d.get('rootDtProjectUuid'), 'bomVersion=', d.get('rootBomVersion'), 'serial=', d.get('rootBomSerialNumber'))")"

echo "==> Fetching branch graph"
GRAPH="$(curl -sS "${RC_BASE_URL}/api/v1/branches/${MAIN_ID}/current/graph")"
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
    print(f\"  edge: {src.get('projectName')} -> {dst.get('projectName')} (bomRef={e.get('bomRef')})\")
"

echo
echo "Done."
echo "  UI: http://localhost:3000/branches/${MAIN_ID}"
echo "  Top project: ${TOP_UUID}"
echo "  Sub project: ${SUB_UUID}"
