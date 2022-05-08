# This builds Docker images similar to the GitHub Actions Virtual Environments,
# with the dependencies we need for end-to-end (e2e) tests.
#
# The runtime user `runner` is setup the same as GitHub Actions also. Notably,
# this allows passwordless `sudo` for RPM and Debian testing.
#
# `make e2e` requires make and go, but `CGO=0` means gcc isn't needed. Ubuntu
# installs more packages, notably windows, for the `check` and `dist` targets.
# To run RPM tests on CentOS, you must build them first on Ubuntu.
#
# This build is intended for use in a matrix, testing all major Linux platforms
# supported by Envoy: Ubuntu and CentOS * amd64 and arm64. Notably, this adds
# CentOS and arm64 which aren't available natively on GitHub Actions. It is
# intended to run arm64 with Travis (as opposed to via emulation). In any case,
# all matrixes should be pushed for local debugging.
#
# Ex. Build the images:
# ```bash
# for parent_image in ubuntu:20.04 centos:8; do docker buildx build \
#                --platform linux/amd64 \
#                --build-arg parent_image=${parent_image} \
#                --build-arg go_stable_release=1_18 \
#                --build-arg go_stable_revision=1.18.1 \
#                --build-arg go_prior_release=1_17 \
#                --build-arg go_prior_revision=1.17.9 \
#                -t func-e-internal:${parent_image//:/-} .github/workflows; done
# ```
#
# Ex. Build func-e on Ubuntu, then end-to-end test on CentOS
# ```bash
# $ docker run --rm -v $PWD:/work func-e-internal:ubuntu-20.04 dist
# $ docker run --rm -v $PWD:/work func-e-internal:centos-8 -o build/func-e_linux_amd64/func-e e2e
# ```
#
# You can troubleshoot like this:
# ```bash
# $ docker run --rm -v $PWD:/work -it --entrypoint /bin/bash func-e-internal:centos-8
# ```
ARG parent_image=centos:8

# This section looks odd, but it is needed to match conventions of the GitHub
# Actions runner. For example, TARGETARCH in Docker is "amd64" whereas GitHub
# actions uses "x64". Moreover, depending on use, case format will change.
# Docker lacks variable substitution options to do this, so we fake it with
# stages. See https://github.com/moby/moby/issues/42904
FROM $parent_image as base-amd64
ARG arch=X64
ARG arch_lc=x64

ARG LINUX
FROM $parent_image as base-arm64
ARG arch=ARM64
ARG arch_lc=arm64

FROM base-${TARGETARCH}

# CentOS runs e2e, but can't run dist as Windows packages are not available.
# While it is possible to build osslsigncode on CentOS, msitools can't due to
# missing libgcab1-devel package. The workaround is to `make dist` with Ubuntu.
ARG centos_packages="make sudo"
# Ubuntu runs check, dist, and e2e, so needs more packages.
ARG ubuntu_packages="make sudo curl git zip wixl msitools osslsigncode"
RUN if [ -f /etc/centos-release ]; then \
    # Change mirrors to vault.centos.org because CentOS 8 went EOL.
    sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* && \
    sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-* && \
    # Use Dandified YUM on CentOS >=8.
    dnf="dnf -qy" && ${dnf} install ${centos_packages} && ${dnf} clean all; \
    else \
    # Use noninteractive to prevent hangs asking about timezone on Ubuntu.
    export DEBIAN_FRONTEND=noninteractive && apt_get="apt-get -qq -y" && \
    ${apt_get} update && ${apt_get} install ${ubuntu_packages} && ${apt_get} clean; \
    fi

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETARCH

# This installs two GOROOTs: the stable and prior release. Two allows pull
# requests to update from a stale release to current without a chicken-egg
# problem or the skew and install time risks of on-demand installation. This
# publishes only two versions as more would bloat the image.
#
# Build args control the precise GOROOTs to install, and should be taken from
# the current GitHub Actions runner. Doing so allows version coherency between
# normal runners and Docker, with version skew bounded by image push frequency.
# See https://github.com/actions/virtual-environments for current versions.
#
# go_XXX_release is the underscore delimited release version. Ex. "1_17"
# go_XXX_revision is the full version number. Ex. "1.17.1"
#
# These are used along with the architecture to build GOROOT variables.
# Ex. GOROOT_1_17_X64=/opt/hostedtoolcache/go/1.17.1/x64
ARG go_stable_revision
ARG go_stable_url=https://golang.org/dl/go${go_stable_revision}.linux-${TARGETARCH}.tar.gz
ARG goroot_stable=${runner_tool_cache}/go/${go_stable_revision}/${arch_lc}
RUN mkdir -p ${goroot_stable} && curl -sSL ${go_stable_url} | tar --strip-components 1 -C ${goroot_stable} -xzpf -

# Dockerfile doesn't support iteration, so repeat above for the prior release.
ARG go_prior_revision
ARG go_prior_url=https://golang.org/dl/go${go_prior_revision}.linux-${TARGETARCH}.tar.gz
ARG goroot_prior=${runner_tool_cache}/go/${go_prior_revision}/${arch_lc}
RUN mkdir -p ${goroot_prior} && curl -sSL ${go_prior_url} | tar --strip-components 1 -C ${goroot_prior} -xzpf -

# Add and switch to the same user as the GitHub Actions runner. This prevents
# ownership problems writing to volumes from the host to docker and visa versa.
ARG user=runner
ARG uid=1001
ARG gid=121
RUN groupadd -f -g ${gid} docker && \
    useradd -u ${uid} -g ${gid} -md /home/runner -s /bin/bash -N ${user} && \
    echo "${user} ALL=NOPASSWD: ALL" >> /etc/sudoers
USER ${user}

# Setup ENV variables used in make that match the GitHub Actions runner.
ENV RUNNER_TOOL_CACHE ${runner_tool_cache}
ARG go_stable_release
ENV GOROOT_${go_stable_release}_${arch} ${goroot_stable}
ARG go_prior_release
ENV GOROOT_${go_prior_release}_${arch} ${goroot_prior}

# Disable gcc to avoid a build dependency on gcc: its glibc might affect Envoy.
ENV CGO_ENABLED 0

# Set CWD to the work directory to avoid overlaps with $HOME.
WORKDIR /work

# Almost everything uses make, but you can override `--entrypoint /bin/bash`.
ENTRYPOINT ["/usr/bin/make"]
CMD ["help"]
