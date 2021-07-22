#!/bin/sh -ue
#
# This script extracts "func-e" for a specific release tag. The release can be a draft.
# Ex. GITHUB_TOKEN=your_repo_token .github/untar_func-e_release.sh v0.3.0

# Crash early on any missing prerequisites
tag_name=$1
curl --version >/dev/null
go version >/dev/null
jq --version >/dev/null
tar --version >/dev/null

# strip the v off the tag name more shell portable than ${tag_name:1}
version=$(echo "${tag_name}" | cut -c2-100) || exit 1

# form the asset name you'd find on the release page
tarball="func-e_${version}_$(go env GOOS)_$(go env GOARCH).tar.gz" || exit 1
curl="curl -fsSL"

# Lookup the last 10 releases, knowing the one we are looking for may not be published.
# See https://docs.github.com/en/rest/reference/repos#list-releases
echo "looking for release that contains ${tarball}"
tarball_url=$(${curl} -H "Authorization: token ${GITHUB_TOKEN}" https://api.github.com/repos/tetratelabs/func-e/releases?per_page=10 |
  jq -er ".|first|.assets|map(select(.name == \"${tarball}\"))|first|.url") || exit 1

# Extract func-e to the current directory per https://docs.github.com/en/rest/reference/repos#get-a-release-asset
echo "extracting func-e from ${tarball_url}"
${curl} -H "Authorization: token ${GITHUB_TOKEN}" -H'Accept: application/octet-stream' "${tarball_url}" | tar -xzf - func-e
./func-e -version
