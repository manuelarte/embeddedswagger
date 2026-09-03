import json
import os
import re
import ssl
import sys
from urllib.error import URLError
from urllib.request import Request, urlopen


def get_project_swagger_version(root: str) -> str:
    path = os.path.join(root, "embeddedswagger.go")
    with open(path, encoding="utf-8") as handle:
        content = handle.read()

    match = re.search(r'SwaggerVersion\s*=\s*"([^"]+)"', content)
    if match is None:
        raise SystemExit("error: could not find SwaggerVersion in embeddedswagger.go")

    version = match.group(1).strip().lstrip("v")
    if not version:
        raise SystemExit("error: invalid SwaggerVersion in embeddedswagger.go")

    return version


def get_github_latest_version() -> str:
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

    return version


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


def main():
    root = sys.argv[1] if len(sys.argv) > 1 else os.getcwd()
    local_version = get_project_swagger_version(root)
    latest_version = get_github_latest_version()

    try:
        local_key = version_key(local_version)
        latest_key = version_key(latest_version)
    except ValueError as exc:
        print(f"error: {exc}")
        raise SystemExit(2)

    print(f"Swagger UI release check: using {local_version} (local), latest upstream is {latest_version}")
    if local_key >= latest_key:
        print("Result: you are not behind the upstream Swagger UI release.")
        raise SystemExit(0)

    print("Result: your embedded Swagger UI is older than the latest upstream release.")
    raise SystemExit(1)


if __name__ == "__main__":
    main()
