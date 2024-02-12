FROM ghcr.io/syntasso/kratix-pipeline-utility:v0.0.1
RUN [ "mkdir", "/scripts" ]

ENV PATH="${PATH}:/scripts/"
COPY resource-configure /scripts/resource-configure
COPY database-configure /scripts/database-configure

CMD [ "sh", "-c", "./resource-configure" ]
ENTRYPOINT []
