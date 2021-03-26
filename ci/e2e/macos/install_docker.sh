#!/usr/bin/env bash

# Below is a copy of https://github.com/play-with-go/play-with-go/blob/d2a13db0ed4ac80b39ce727cd54f2438c93096dc/_scripts/macCISetup.sh
# The last change to this above was in response to this comment https://github.com/docker/for-mac/issues/2359#issuecomment-793595407
# Since then, this incorporates a different version of 'Docker for Mac' in attempts to speed execution.

set -euo pipefail

# We can't have this line because the default version of bash on mac os too old
# shopt -s inherit_errexit

# With thanks/credit to https://github.com/docker/for-mac/issues/2359#issuecomment-607154849

# Update brew to make sure we're using the latest formulae
brew update

###############################
# General
brew install bash
brew install gsed
brew install findutils
ln -s /usr/local/bin/gsed /usr/local/bin/sed
ln -s /usr/local/bin/gfind /usr/local/bin/find
hash -r
which sed
which find

###############################
# Docker

# Install Docker
TMP_DIR=$(mktemp -d)

# Docker for Mac 2.0.0.3-ce-mac81,31259 (the last version of 'Docker for Mac' that can be installed in CI environment)
E2E_MACOS_DOCKER_CASK_VERSION="${E2E_MACOS_DOCKER_CASK_VERSION:-8ce4e89d10716666743b28c5a46cd54af59a9cc2}"

# install Docker for Mac
pushd "${TMP_DIR}"
curl -L https://raw.githubusercontent.com/Homebrew/homebrew-cask/${E2E_MACOS_DOCKER_CASK_VERSION}/Casks/docker.rb > docker.rb
brew install --cask docker.rb
popd

# Allow the app to run without confirmation
xattr -d -r com.apple.quarantine /Applications/Docker.app

# preemptively do docker.app's setup to avoid any gui prompts
sudo /bin/cp /Applications/Docker.app/Contents/Library/LaunchServices/com.docker.vmnetd /Library/PrivilegedHelperTools
sudo /bin/cp /Applications/Docker.app/Contents/Resources/com.docker.vmnetd.plist /Library/LaunchDaemons/
sudo /bin/chmod 544 /Library/PrivilegedHelperTools/com.docker.vmnetd
sudo /bin/chmod 644 /Library/LaunchDaemons/com.docker.vmnetd.plist
sudo /bin/launchctl load /Library/LaunchDaemons/com.docker.vmnetd.plist

# Run
[[ $(uname) == 'Darwin' ]] || {
	echo "This function only runs on macOS." >&2
	exit 2
}

echo "-- Starting Docker.app, if necessary..."

open -g -a /Applications/Docker.app || exit

# Wait for the server to start up, if applicable.
i=0
while ! docker system info &> /dev/null; do
	((i++ == 0)) && printf %s '-- Waiting for Docker to finish starting up...' || printf '.'
	sleep 1
done
((i))   && printf '\n'

echo "-- Docker is ready."
