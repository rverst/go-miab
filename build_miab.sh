#! /bin/bash

VERSION="1.0.0"

#The build script is designed to run in the official golang docker container, which does not have 'zip' installed.
#So we just install it.
command -v zip >/dev/null 2>&1 || { apt-get update && apt-get -y install zip; }
#command -v zip >/dev/null 2>&1 || { echo "Scrips need 'zip' to pack the windows binaries."; exit 1; }

DIR=$PWD

BUILDDIR="${DIR}/build/${VERSION}"
CHECKSUM="${BUILDDIR}/miab_${VERSION}_checksums.txt"
LICENSE="${DIR}/LICENSE.txt"
README="${DIR}/README.md"
CONFIG="${DIR}/test_config.yaml"


if [[ -d "$BUILDDIR" ]]; then
  rm -r "$BUILDDIR"
fi

 mkdir -p "$BUILDDIR"

echo "$DIR"
echo "$BUILDDIR"

touch "$CHECKSUM"
echo "sha256 checksums" > "$CHECKSUM"

for GOOS in windows darwin linux; do
  for GOARCH in 386 amd64; do
    EXT=""
    if [[ $GOOS == "windows" ]]; then
      EXT=".exe"
    fi
    OSARCH="miab_${GOOS}-${GOARCH}"
    OSARCH_BUILDDIR="${BUILDDIR}/${OSARCH}"
    FILE="${OSARCH_BUILDDIR}/miab${EXT}"

    export GOOS GOARCH
    mkdir -p "$OSARCH_BUILDDIR"

    go build -v -o "${FILE}" ./cmd/cli/miab.go

    cd "$BUILDDIR" || exit 1

    cp "$LICENSE" "${OSARCH_BUILDDIR}/LICENSE.txt"
    cp "$README" "${OSARCH_BUILDDIR}/README.md"
    cp "$CONFIG" "${OSARCH_BUILDDIR}/test_config.yaml"

    if [[ $GOOS == "windows" ]]; then
      AFILE="miab_${VERSION}_${GOOS}-${GOARCH}.zip"
      ARCHIVE="${BUILDDIR}/${AFILE}"
      zip -r "$ARCHIVE" "$OSARCH"
      sha256sum "$AFILE" >> "$CHECKSUM"

    else
      AFILE="miab_${VERSION}_${GOOS}-${GOARCH}.tar.gz"
      ARCHIVE="${BUILDDIR}/${AFILE}"
      tar -czvf "$ARCHIVE" "$OSARCH"
      sha256sum "$AFILE" >> "$CHECKSUM"
    fi

    cd "$DIR" || exit 1
  done
done
