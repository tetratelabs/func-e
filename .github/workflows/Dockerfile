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
# $ for parent_image in ubuntu:20.04 centos:8; do docker build \
#                --build-arg parent_image=${parent_image} \
#                --build-arg go_revision=1.17.1 \
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
# $ docker run --rm -v $PWD:/work -it --entrypoint --entrypoint /bin/bash func-e-internal:centos-8
# ```
ARG parent_image=centos:8
FROM ${parent_image}

# CentOS runs e2e, but can't run dist as Windows packages are not available.
# While it is possible to build osslsigncode on CentOS, msitools can't due to
# missing libgcab1-devel package. The workaround is to `make dist` with Ubuntu.
ARG centos_packages="make sudo"
# Ubuntu runs check, dist, and e2e, so needs more packages.
ARG ubuntu_packages="make sudo curl git zip wixl msitools osslsigncode"
RUN if [ -f /etc/centos-release ]; then \
    # Use Dandified YUM on CentOS >=8. \
    dnf="dnf -qy" && ${dnf} install ${centos_packages} && ${dnf} clean all; \
    else \
    # Use noninteractive to prevent hangs asking about timezone on Ubuntu. \
    export DEBIAN_FRONTEND=noninteractive && apt_get="apt-get -qq -y" && \
    ${apt_get} update && ${apt_get} install ${ubuntu_packages} && ${apt_get} clean; \
    fi

# See https://docs.docker.com/engine/reference/builder/#automatic-platform-args-in-the-global-scope
ARG TARGETARCH
ARG go_revision
ARG go_stable_url=https://golang.org/dl/go${go_revision}.linux-${TARGETARCH}.tar.gz

# Extract into the default GOROOT
RUN curl -sSL ${go_stable_url} | tar -xzC /usr/local

# Add and switch to the same user as the GitHub Actions runner. This prevents
# ownership problems writing to volumes from the host to docker and visa versa.
ARG user=runner
ARG uid=1001
ARG gid=121
RUN groupadd -f -g ${gid} docker && \
    useradd -u ${uid} -g ${gid} -md /home/runner -s /bin/bash -N ${user} && \
    echo "${user} ALL=NOPASSWD: ALL" >> /etc/sudoers
USER ${user}

# Setup the default GOROOT
ENV GOROOT=/usr/local/go
ENV PATH=$GOROOT/bin:$PATH

# Disable gcc to avoid a build dependency on gcc: its glibc might affect Envoy.
ENV CGO_ENABLED 0

# Set CWD to the work directory to avoid overlaps with $HOME.
WORKDIR /work

# Almost everything uses make, but you can override `--entrypoint /bin/bash`.
ENTRYPOINT ["/usr/bin/make"]
CMD ["help"]
