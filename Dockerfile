FROM gcr.io/distroless/cc

COPY bin/getenvoy /

COPY bin/linux_glibc /root/.getenvoy/builds/standard/1.11.1/linux_glibc

# Reference is hardcoded for now as I don't think theres a way around this.
# We may have to use bazel to build our Docker images...
ENTRYPOINT ["/getenvoy", "run", "standard:1.11.1/linux-glibc"]

