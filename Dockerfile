FROM alpine:3.7
COPY trovilo /bin/
ENTRYPOINT ["/bin/trovilo"]
CMD ["--config /config.yaml", "--log-json"]
