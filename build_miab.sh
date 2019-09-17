#! /bin/bash

#The build script is designed to run in the official golang docker container, which does not have 'zip' installed.
#So we just install it. If you use the scipt otherwise, disable the installation and exit the script at this stage.
function checkDependencies() {
  if [[ ! $(command -v zip) ]]; then
    printf "This script needs 'zip' to pack the windows binaries.\n\n"
    apt-get update && apt-get -y install zip
    #exit 1
  fi
}

# check for semantic version, see https://gist.github.com/rverst/1f0b97da3cbeb7d93f4986df6e8e5695
function checkVersion() {
  if [[ $1 =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(\+([0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*))?$ ]]; then
    echo "$1"
  else
    echo ""
  fi
}

function checkVersionEx() {
  if [[ $1 =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(\+([0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*))?$ ]]; then
    echo "${1:1}"
  else
    echo "$(checkVersion ${1})"
  fi
}

function main() {

  COMHASH="unknown"
  if [[ $1 ]]; then
    VERSION=$(checkVersionEx "$1")
    if [[ ! $VERSION ]]; then
      printf "The provided version is in the wrong format, please provide as semantic version.\n\n"
      exit 1
    fi
  fi

  if [[ $(command -v git) ]]; then
    COMHASH=$(git rev-parse HEAD)

    if [[ ! $VERSION ]]; then
      TAGS=$(git tag -l 'v*' --points-at "$COMHASH")
      VERSION=$(checkVersion "$TAGS")
      if [[ ! $VERSION ]]; then
        printf "Can't find a valid version tag on HEAD, please provide the version as parameter.\n"
        printf  "Either there is no valid tag (e.g. v1.0.0) or there is more than one.\n"
        if [[ $TAGS ]]; then
          printf  "Tags found:\n"
          printf  "%s\n\n" "$TAGS"
        fi
        exit 1
      fi
    fi
  fi

  PACKAGE="github.com/rverst/go-miab/cmd/cli"
  BUILDDATE=$(date -u -Iseconds)
  LDFLAGS="-X $PACKAGE/command.Version=$VERSION -X $PACKAGE/command.CommitHash=$COMHASH -X $PACKAGE/command.BuildDate=$BUILDDATE"

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
  echo "sha256 checksums" >"$CHECKSUM"

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

      go build -v -ldflags "$LDFLAGS" -o "$FILE" ./cmd/cli/miab.go

      cd "$BUILDDIR" || exit 1

      cp "$LICENSE" "${OSARCH_BUILDDIR}/LICENSE.txt"
      cp "$README" "${OSARCH_BUILDDIR}/README.md"
      cp "$CONFIG" "${OSARCH_BUILDDIR}/test_config.yaml"

      if [[ $GOOS == "windows" ]]; then
        AFILE="miab_${VERSION}_${GOOS}-${GOARCH}.zip"
        ARCHIVE="${BUILDDIR}/${AFILE}"
        zip -r "$ARCHIVE" "$OSARCH"
        sha256sum "$AFILE" >>"$CHECKSUM"

      else
        AFILE="miab_${VERSION}_${GOOS}-${GOARCH}.tar.gz"
        ARCHIVE="${BUILDDIR}/${AFILE}"
        tar -czvf "$ARCHIVE" "$OSARCH"
        sha256sum "$AFILE" >>"$CHECKSUM"
      fi

      cd "$DIR" || exit 1
    done
  done

}

main "$1"
