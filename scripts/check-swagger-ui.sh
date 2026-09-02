#!/usr/bin/env bash
set -eu

ROOT="${1:-$(pwd)}"
SOURCE_PATH="${ROOT}/embeddedswagger.go"

if [[ ! -f "${SOURCE_PATH}" ]]; then
  echo "error: unable to read ${SOURCE_PATH}"
  exit 2
fi

LOCAL_VERSION="$(python3 - "${SOURCE_PATH}" <<'PY'
import re
import sys

path = sys.argv[1]
content = open(path, encoding="utf-8").read()
match = re.search(r'SwaggerVersion\s*=\s*"([^"]+)"', content)
if match is None:
    raise SystemExit("error: could not find SwaggerVersion in embeddedswagger.go")

version = match.group(1).strip().lstrip("v")
if not version:
    raise SystemExit("error: invalid SwaggerVersion in embeddedswagger.go")

print(version)
PY
)"

LATEST_VERSION="$(python3 - <<'PY'
import json
import ssl
from urllib.error import URLError
from urllib.request import Request, urlopen

req = Request(
    "https://api.github.com/repos/swagger-api/swagger-ui/releases/latest",
    headers={
        "Accept": "application/vnd.github+json",
        "User-Agent": "embeddedswagger-check",
    },
)

try:
    with urlopen(req, timeout=20, context=ssl.create_default_context()) as response:
        payload = json.load(response)
except URLError:
    with urlopen(req, timeout=20, context=ssl._create_unverified_context()) as response:
        payload = json.load(response)

version = str(payload.get("tag_name", "")).strip().lstrip("v")
if not version:
    raise SystemExit("error: GitHub API did not return a valid tag_name for the latest release")

print(version)
PY
)"

python3 - "${LOCAL_VERSION}" "${LATEST_VERSION}" <<'PY'
import re
import sys

LOCAL_VERSION = sys.argv[1].strip().lstrip("v")
LATEST_VERSION = sys.argv[2].strip().lstrip("v")


def version_key(version: str):
    version = version.strip().lstrip("v")
    match = re.match(r"^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([0-9A-Za-z.-]+))?$", version)
    if match is None:
        raise ValueError(f"unsupported version format: {version!r}")

    major, minor, patch, prerelease = match.groups()
    return (
        int(major or 0),
        int(minor or 0),
        int(patch or 0),
        0 if prerelease is None else -1,
        prerelease or "",
    )

try:
    local_key = version_key(LOCAL_VERSION)
    latest_key = version_key(LATEST_VERSION)
except ValueError as exc:
    print(f"error: {exc}")
    raise SystemExit(2)

print(f"Swagger UI release check: using {LOCAL_VERSION} (local), latest upstream is {LATEST_VERSION}")
if local_key >= latest_key:
    print("Result: you are not behind the upstream Swagger UI release.")
    raise SystemExit(0)

print("Result: your embedded Swagger UI is older than the latest upstream release.")
raise SystemExit(1)
PY
