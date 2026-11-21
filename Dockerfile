# 源码构建只需要 Go 与可信公钥，组件从公开签名 Release 获取。
FROM --platform=$BUILDPLATFORM golang:1.27.1-bookworm@sha256:8d48e12ec56735e9358640898b9d9b9fcca110612ed8a5567438c0a1baa24e66 AS build
WORKDIR /src
ENV CGO_ENABLED=0
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
RUN : "${VERSION:?VERSION is required}" && \
    go run ./scripts/dist --version "$VERSION" --public-key .github/release.pub --out /dist --platforms "$TARGETOS/$TARGETARCH" && \
    mkdir /image && \
    tar -xzf "/dist/alonabot-$VERSION-$TARGETOS-$TARGETARCH.tar.gz" -C /image --strip-components=1 && \
    mkdir /image/data

# published 消费已验收的发行归档，发布两架构时复用同一批产物。
FROM scratch AS runtime
ENV ALONABOT_CONTAINER=1 PATH=/
USER 10001:10001
VOLUME /data
ENTRYPOINT ["/alonabot"]
CMD ["run", "--data-dir", "/data"]

FROM runtime AS published
COPY --from=distribution /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=distribution /alonabot /alonabot
COPY --from=distribution /packages/ /packages/
COPY --from=distribution /LICENSE /LICENSE
COPY --from=distribution /source.json /source.json
COPY --from=distribution --chown=10001:10001 /data/ /data/

FROM runtime AS source
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /image/alonabot /alonabot
COPY --from=build /image/packages/ /packages/
COPY --from=build --chown=10001:10001 /image/data/ /data/
COPY --from=build /image/LICENSE /LICENSE
COPY --from=build /image/source.json /source.json
