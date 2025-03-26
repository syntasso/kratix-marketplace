FROM ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1
RUN [ "mkdir", "/scripts" ]

ENV PATH="${PATH}:/scripts/"
COPY scripts/ /scripts/

CMD [ "sh", "-c", "./resource-configure" ]
ENTRYPOINT []
