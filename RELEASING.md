# Releasing

One version, one place. The tag in the repository, the bundle in the release
archive and the page on GitHub carry the same number.

## 1. Build

```bash
make release            # release/: four binaries, ui/, module.yaml
```

## 2. Sign

The bundle carries `checksums.txt` over every file and a signature over that
list, made with the key in `~/harness/maintenance-signing`.

```bash
cd release
find . -type f ! -name 'checksums.txt*' | sort | sed 's|^\./||' | xargs sha256sum > checksums.txt
openssl dgst -sha256 -sign ~/harness/maintenance-signing/private_key.pem \
  -out /tmp/sig.bin checksums.txt
base64 -w0 /tmp/sig.bin > checksums.txt.sig
openssl dgst -sha256 -verify ~/harness/maintenance-signing/public_key.pem \
  -signature /tmp/sig.bin checksums.txt
tar czf /tmp/maintenance-<version>.tar.gz *
```

The Windows zip holds the executable alone, as it always did:

```bash
python3 - <<'PY'
import zipfile, pathlib
src = pathlib.Path('maintenance_windows_amd64.exe')
with zipfile.ZipFile('/tmp/maintenance-<version>-windows-amd64.zip', 'w',
                     zipfile.ZIP_DEFLATED, compresslevel=6) as z:
    z.write(src, arcname=src.name)
PY
```

## 3. Store

Everything belongs in `~/harness/maintenance-release/<version>/`:

```
<version>/
  maintenance-<version>.tar.gz
  maintenance-<version>-windows-amd64.zip
  build/                    the unpacked release directory, as it was built
```

`release/` in the project is a working directory: `make release` wipes it. What
is worth keeping goes to the archive before the next build.

## 4. Tag and release

```bash
git tag -a v<version> -m "Maintenance <version>"
git push origin main
git push origin v<version>
```

Then create the GitHub release from that tag with `RELEASE-<version>.md` as its
body. Publishing it hands the commit to the marketplace builder, which builds
every platform and the bundle itself, and a person at Prusa reviews it.

## 5. The marketplace page

Images live in `artwork/`:

```
marketplace.png   2048x2048, under 5 MB, the listing's logo
card.png          2741x1530, the tile picture, uploaded first
```

## 6. Check before saying it is done

```bash
git status --porcelain                       # empty
git rev-list --count origin/main..HEAD       # 0
git rev-list -n1 v<version>                  # equals HEAD
grep -m1 '^version:' module.yaml             # equals the tag
grep -m1 '^## ' CHANGELOG.md                 # equals the version
```

A tag pointing at a commit `main` does not have is the mistake this list exists
for.
