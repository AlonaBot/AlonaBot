# 主机工具链调用发行构建，为目标平台生成签名组件与 scratch 镜像。
FROM --platform=$BUILDPLATFORM node:26.5.1-bookworm@sha256:057896e7e3f542ff82f7e65f106ac04e646775316449e676dfe8fb053d3b43ef AS node
FROM --platform=$BUILDPLATFORM golang:1.27.1-bookworm@sha256:8d48e12ec56735e9358640898b9d9b9fcca110612ed8a5567438c0a1baa24e66 AS build

COPY --from=node /usr/local/ /usr/local/
WORKDIR /src
ENV CGO_ENABLED=0
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG ASSET_BASE

RUN --mount=type=secret,id=release_key,required=true \
    : "${VERSION:?VERSION is required}" && \
    set -- --version "$VERSION" --key /run/secrets/release_key --out /dist --platforms "$TARGETOS/$TARGETARCH" && \
    if [ -n "$ASSET_BASE" ]; then set -- "$@" --asset-base "$ASSET_BASE"; fi && \
    go run ./scripts/dist "$@" && \
    mkdir /image && \
    tar -xzf "/dist/alonabot-$VERSION-$TARGETOS-$TARGETARCH.tar.gz" -C /image --strip-components=1 && \
    mkdir /image/data

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /image/alonabot /alonabot
COPY --from=build /image/packages/ /packages/
COPY --from=build --chown=10001:10001 /image/data/ /data/
ENV ALONABOT_CONTAINER=1 PATH=/
USER 10001:10001
VOLUME /data
ENTRYPOINT ["/alonabot"]
CMD ["run", "--data-dir", "/data"]
