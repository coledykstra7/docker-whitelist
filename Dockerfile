# Squid Dockerfile for whitelist proxy
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y squid && rm -rf /var/lib/apt/lists/*
COPY squid/squid.conf /etc/squid/squid.conf
EXPOSE 3128
VOLUME ["/data"]
CMD ["squid", "-N", "-f", "/etc/squid/squid.conf"]
