FROM ollama/ollama:latest

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Pre-pull TinyDolphin at build time
RUN /bin/sh -c 'set -e; ollama serve & pid=$!; \
  for i in $(seq 1 120); do curl -fsS http://127.0.0.1:11434/api/version && break || sleep 1; done; \
  ollama pull tinydolphin:latest; \
  kill -TERM "$pid"; wait "$pid" || true'

ENV OLLAMA_HOST=0.0.0.0
EXPOSE 11434

COPY start.sh /usr/local/bin/start.sh
RUN chmod +x /usr/local/bin/start.sh
ENTRYPOINT ["/usr/local/bin/start.sh"]
