# Build app
FROM rust:latest as builder
WORKDIR /usr/src/speedtest-exporter
COPY . .
RUN --mount=type=cache,target=/usr/local/cargo/registry \
    --mount=type=cache,target=/usr/src/speedtest-exporter/target \
    cargo install --path .

# Download speedtest CLI
FROM ubuntu:latest as download-speedtest
RUN apt-get update && \
    apt-get install -y curl
WORKDIR /tmp
RUN curl -O "https://install.speedtest.net/app/cli/ookla-speedtest-1.2.0-linux-aarch64.tgz"
RUN tar xvf *.tgz

# Build target image
FROM ubuntu:latest
RUN apt-get update && \
    apt-get install -y openssl ca-certificates && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /usr/local/cargo/bin/speedtest-exporter /usr/local/bin/speedtest-exporter
COPY --from=download-speedtest /tmp/speedtest /usr/local/bin
ENTRYPOINT ["speedtest-exporter"]
