import os
import re
import shutil
import stat
import sys
import tempfile
import zipfile
from pathlib import Path, PurePosixPath
from urllib.parse import quote
from urllib.request import Request, urlopen

from main import get_github_latest_tag, get_project_swagger_version, version_key


VERSION_PATTERN = re.compile(r'(SwaggerVersion\s*=\s*")([^"]+)(")')
REQUIRED_FILES = {
    "index.html",
    "swagger-initializer.js",
    "swagger-ui-bundle.js",
    "swagger-ui-standalone-preset.js",
    "swagger-ui.css",
}
INITIALIZER_URL = "https://petstore.swagger.io/v2/swagger.json"


def extract_dist(archive: Path, destination: Path) -> None:
    with zipfile.ZipFile(archive) as source:
        roots = set()
        found = set()
        for entry in source.infolist():
            path = PurePosixPath(entry.filename)
            if len(path.parts) < 3 or path.parts[1] != "dist":
                continue
            roots.add(path.parts[0])
            relative = path.parts[2:]
            if any(part in (".", "..") for part in relative) or "\\" in entry.filename:
                raise ValueError(f"unsafe distribution path: {entry.filename}")
            if stat.S_ISLNK(entry.external_attr >> 16):
                raise ValueError(f"distribution contains a symbolic link: {entry.filename}")
            target = destination.joinpath(*relative)
            if entry.is_dir():
                target.mkdir(parents=True, exist_ok=True)
            else:
                target.parent.mkdir(parents=True, exist_ok=True)
                with source.open(entry) as src, target.open("xb") as dst:
                    shutil.copyfileobj(src, dst)
                found.add("/".join(relative))

    if len(roots) != 1 or not REQUIRED_FILES.issubset(found):
        raise ValueError("archive does not contain a complete Swagger UI dist directory")
    initializer = (destination / "swagger-initializer.js").read_text(encoding="utf-8")
    if initializer.count(INITIALIZER_URL) != 1:
        raise ValueError("upstream initializer no longer contains the expected OpenAPI URL exactly once")


def update(root: Path, tag: str, archive: Path) -> None:
    version = tag.removeprefix("v")
    go_file = root / "embeddedswagger.go"
    content = go_file.read_text(encoding="utf-8")
    updated, count = VERSION_PATTERN.subn(lambda match: match[1] + version + match[3], content)
    if count != 1:
        raise ValueError("expected exactly one SwaggerVersion assignment in embeddedswagger.go")

    static = root / "static"
    with tempfile.TemporaryDirectory(prefix=".swagger-update-", dir=static) as temp:
        temp_path = Path(temp)
        dist = temp_path / "dist"
        dist.mkdir()
        extract_dist(archive, dist)
        replacement = temp_path / "embeddedswagger.go"
        replacement.write_text(updated, encoding="utf-8")
        replacement.chmod(go_file.stat().st_mode)

        target = static / "swagger-ui"
        backup = temp_path / "previous-dist"
        target.rename(backup)
        try:
            dist.rename(target)
            os.replace(replacement, go_file)
        except OSError:
            if target.exists():
                shutil.rmtree(target)
            backup.rename(target)
            raise


def main() -> None:
    root = Path(sys.argv[1] if len(sys.argv) > 1 else os.getcwd()).resolve()
    local_version = get_project_swagger_version(str(root))
    tag = get_github_latest_tag()
    version = tag.removeprefix("v")
    if version_key(local_version) >= version_key(version):
        print(f"Swagger UI is up to date: using {local_version}, latest upstream is {version}")
        return

    print(f"Updating Swagger UI from {local_version} to {version}")
    url = f"https://github.com/swagger-api/swagger-ui/archive/refs/tags/{quote(tag, safe='')}.zip"
    request = Request(url, headers={"User-Agent": "embeddedswagger-update"})
    with tempfile.TemporaryDirectory(prefix="swagger-download-") as temp:
        archive = Path(temp) / "release.zip"
        with urlopen(request, timeout=60) as response, archive.open("wb") as target:
            shutil.copyfileobj(response, target)
        update(root, tag, archive)
    print(f"Updated Swagger UI to {version}")


if __name__ == "__main__":
    main()
