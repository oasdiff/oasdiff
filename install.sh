#!/bin/sh
set -e

# Adapted from https://raw.githubusercontent.com/railwayapp/cli/master/install.sh
#
# oasdiff
# https://github.com/oasdiff/oasdiff


INSTALL_DIR=${INSTALL_DIR:-"/usr/local/bin"}
BINARY_NAME=${BINARY_NAME:-"oasdiff"}

REPO_NAME="oasdiff/oasdiff"
ISSUE_URL="https://github.com/oasdiff/oasdiff/issues/new"

get_latest_release() {
  curl --silent -L "https://api.github.com/repos/$1/releases/latest" | # Get latest release from GitHub api
    grep '"tag_name":' |                                            # Get tag line
    sed -E 's/.*"([^"]+)".*/\1/'                                    # Pluck JSON value
}

get_asset_name() {
  echo "oasdiff_$1_$2_$3.tar.gz"
}

get_download_url() {
  local asset_name=$(get_asset_name $1 $2 $3)
  echo "https://github.com/oasdiff/oasdiff/releases/download/v$1/${asset_name}"
}

get_checksum_url() {
  echo "https://github.com/oasdiff/oasdiff/releases/download/v$1/checksums.txt"
}

command_exists() {
  command -v "$@" >/dev/null 2>&1
}

get_os() {
  case "$(uname -s)" in
    *linux* ) echo "linux" ;;
    *Linux* ) echo "linux" ;;
    *darwin* ) echo "darwin" ;;
    *Darwin* ) echo "darwin" ;;
  esac
}

get_machine() {
  case "$(uname -m)" in
    "x86_64"|"amd64"|"x64")
      echo "amd64" ;;
    "i386"|"i86pc"|"x86"|"i686")
      echo "i386" ;;
    "arm64"|"armv6l"|"aarch64")
      echo "arm64"
  esac
}

get_tmp_dir() {
  echo $(mktemp -d)
}

do_checksum() {
  echo "Validating checksum"
  checksum_url=$(get_checksum_url $version)
  get_checksum_url $version
  expected_checksum=$(curl -sL $checksum_url | grep $asset_name | awk '{print $1}')

  if command_exists sha256sum; then
    checksum=$(sha256sum $asset_name | awk '{print $1}')
  elif command_exists shasum; then
    checksum=$(shasum -a 256 $asset_name | awk '{print $1}')
  else
    echo "Could not find a checksum program. Install shasum or sha256sum to validate checksum."
    return 0
  fi

  if [ "$checksum" != "$expected_checksum" ]; then
    echo "Checksums do not match"
    exit 1
  fi
}

do_install() {
  asset_name=$(get_asset_name $version $os $machine)
  download_url=$(get_download_url $version $os $machine)

  command_exists curl || {
    echo "curl is not installed"
    exit 1
  }

  command_exists tar || {
    echo "tar is not installed"
    exit 1
  }

  local tmp_dir=$(get_tmp_dir)
  echo "Temporary directory is $tmp_dir"

  echo "Downloading $download_url"
  (cd $tmp_dir && curl -sL -O "$download_url")

  (cd $tmp_dir && do_checksum)

  echo "Extracting tar file"
  (cd $tmp_dir && tar -xzf "$asset_name")

  echo "Installing $BINARY_NAME into $tmp_dir"
  mv "$tmp_dir/$BINARY_NAME" $INSTALL_DIR

  chmod +x $INSTALL_DIR/$BINARY_NAME
  echo "Installed oasdiff to $INSTALL_DIR"

  echo "Removing temporary directory"
  rm -rf $tmp_dir
}

main() {
  if [ -n "$version" ]; then
    version="$version"
  else
    latest_tag=$(get_latest_release $REPO_NAME)
    version=$(echo $latest_tag | sed 's/v//')
  fi

  os=$(get_os)
  if test -z "$os"; then
    echo "$(uname -s) os type is not supported"
    echo "Please create an issue so we can add support. $ISSUE_URL"
    exit 1
  fi

  if [ "$os" = "darwin" ]; then
    machine="all"
  else
    machine=$(get_machine)
  fi

  if test -z "$machine"; then
    echo "$(uname -m) machine type is not supported"
    echo "Please create an issue so we can add support. $ISSUE_URL"
    exit 1
  fi
  
  do_install
  
  echo "oasdiff is now installed! type 'oasdiff -h' to see a list of commands"
}

main