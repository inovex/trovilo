FROM alpine:3.7
RUN \
  mkdir /lib64 && \
  ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 && \
  apk add \
    --no-cache \
    --update \
    curl \
    jq
COPY trovilo /bin/
COPY promtool /bin/
ENTRYPOINT ["/bin/trovilo"]
CMD ["--config", "/config.yaml", "--log-json"]