import tempfile
import unittest
import zipfile
from pathlib import Path
from unittest.mock import patch

import main
import update_swagger_ui


class UpdateSwaggerUITest(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        self.go_file = self.root / "embeddedswagger.go"
        self.go_file.write_text('const SwaggerVersion = "1.0.0"\n', encoding="utf-8")
        self.dist = self.root / "static" / "swagger-ui"
        self.dist.mkdir(parents=True)
        (self.dist / "obsolete.js").write_text("old", encoding="utf-8")
        self.archive = self.root / "release.zip"

    def make_archive(self, *, omit=None, initializer=None, extra=None):
        if initializer is None:
            initializer = update_swagger_ui.INITIALIZER_URL
        with zipfile.ZipFile(self.archive, "w") as archive:
            for name in update_swagger_ui.REQUIRED_FILES - {omit}:
                content = initializer if name == "swagger-initializer.js" else name
                archive.writestr(f"swagger-ui-release/dist/{name}", content)
            archive.writestr("swagger-ui-release/dist/swagger-ui.js.map", "source map")
            if extra:
                archive.writestr(f"swagger-ui-release/dist/{extra}", "unsafe")

    def assert_original_unchanged(self):
        self.assertEqual(self.go_file.read_text(encoding="utf-8"), 'const SwaggerVersion = "1.0.0"\n')
        self.assertEqual(sorted(path.name for path in self.dist.iterdir()), ["obsolete.js"])

    def test_replaces_distribution_and_version(self):
        self.make_archive()
        update_swagger_ui.update(self.root, "v2.3.4", self.archive)
        self.assertEqual(self.go_file.read_text(encoding="utf-8"), 'const SwaggerVersion = "2.3.4"\n')
        self.assertFalse((self.dist / "obsolete.js").exists())
        self.assertTrue((self.dist / "swagger-ui.js.map").is_file())
        self.assertTrue(all((self.dist / name).is_file() for name in update_swagger_ui.REQUIRED_FILES))

    def test_missing_required_file_does_not_change_project(self):
        self.make_archive(omit="index.html")
        with self.assertRaisesRegex(ValueError, "complete"):
            update_swagger_ui.update(self.root, "v2.3.4", self.archive)
        self.assert_original_unchanged()

    def test_changed_initializer_does_not_change_project(self):
        self.make_archive(initializer="another URL")
        with self.assertRaisesRegex(ValueError, "initializer"):
            update_swagger_ui.update(self.root, "v2.3.4", self.archive)
        self.assert_original_unchanged()

    def test_unsafe_path_does_not_change_project(self):
        self.make_archive(extra="../../outside")
        with self.assertRaisesRegex(ValueError, "unsafe"):
            update_swagger_ui.update(self.root, "v2.3.4", self.archive)
        self.assert_original_unchanged()
        self.assertFalse((self.root / "static" / "outside").exists())

    def test_duplicate_version_does_not_change_project(self):
        self.make_archive()
        self.go_file.write_text('SwaggerVersion = "1.0.0"\nSwaggerVersion = "1.0.0"\n', encoding="utf-8")
        with self.assertRaisesRegex(ValueError, "exactly one"):
            update_swagger_ui.update(self.root, "v2.3.4", self.archive)
        self.assertTrue((self.dist / "obsolete.js").exists())

    def test_failed_version_write_restores_distribution(self):
        self.make_archive()
        with patch.object(update_swagger_ui.os, "replace", side_effect=OSError("write failed")):
            with self.assertRaisesRegex(OSError, "write failed"):
                update_swagger_ui.update(self.root, "v2.3.4", self.archive)
        self.assert_original_unchanged()

    def test_up_to_date_skips_download(self):
        with (
            patch.object(update_swagger_ui.sys, "argv", ["update", str(self.root)]),
            patch.object(update_swagger_ui, "get_github_latest_tag", return_value="v1.0.0"),
            patch.object(update_swagger_ui, "urlopen") as download,
        ):
            update_swagger_ui.main()
        download.assert_not_called()
        self.assert_original_unchanged()


class ReleaseCheckTest(unittest.TestCase):
    def test_latest_tag_requires_version_string(self):
        class Response:
            def __enter__(self):
                return self

            def __exit__(self, *args):
                pass

            def read(self):
                return b'{"tag_name": "../invalid"}'

        with patch.object(main, "urlopen", return_value=Response()):
            with self.assertRaisesRegex(SystemExit, "valid tag_name"):
                main.get_github_latest_tag()


if __name__ == "__main__":
    unittest.main()
