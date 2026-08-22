ARG BUILD_FROM=ghcr.io/home-assistant/aarch64-base-debian:bookworm
FROM ${BUILD_FROM}

# Provided automatically by buildx from --platform (amd64 / arm64).
ARG TARGETARCH

# Prebuilt, frontend-embedded dashi binary, staged by CI at dist/<arch>/dashi.
COPY dist/${TARGETARCH}/dashi /usr/bin/dashi
COPY run.sh /run.sh
RUN chmod +x /usr/bin/dashi /run.sh

CMD [ "/run.sh" ]
