FROM caddy:builder AS builder
ENV GOTOOLCHAIN=auto
ARG CADDY_MODULE_PATH=github.com/Bl4cky99/caddy-redirector
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    xcaddy build --with ${CADDY_MODULE_PATH}=/src --output /caddy

FROM caddy:latest
COPY --from=builder /caddy /usr/bin/caddy