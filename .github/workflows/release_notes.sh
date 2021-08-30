#!/bin/sh -ue
#
# This script generates the release notes "func-e" for a specific release tag.
# .github/release_notes.sh v0.3.0

tag=$1
prior_tag=$(git tag -l 'v*'|sed "/${tag}/,+10d"|tail -1)
range="${prior_tag}..${tag}"

git config log.mailmap true
changelog=$(git log --format='%h %s %aN, %(trailers:key=co-authored-by)' "${range}")
contributors=$(git shortlog -s --group=author --group=trailer:co-authored-by "${range}" | cut -f 2 |sort)

# strip the v off the tag name more shell portable than ${tag:1}
version=$(echo "${tag}" | cut -c2-100) || exit 1
cat <<EOF
func-e ${version} supports X and Y and notably fixes Z

TODO: classify the below into up to 4 major headings and the rest as bulleted items in minor changes
The published release notes should only include the summary statement in this section.

${changelog}

## X packages

Don't forget to cite who was involved and why

## func-e Y

## Minor changes

TODO: don't add trivial things like fixing spaces or non-concerns like build glitches

* Z is now fixed thanks to Yogi Bear

## Thank you
func-e ${version} was possible thanks to the following contributors:

${contributors}

EOF
