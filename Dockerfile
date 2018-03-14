FROM alpine:3.7
COPY trovilo /bin/
RUN \
  mkdir /lib64 && \
  ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
ENTRYPOINT ["/bin/trovilo"]
CMD ["--config", "/config.yaml", "--log-json"]
