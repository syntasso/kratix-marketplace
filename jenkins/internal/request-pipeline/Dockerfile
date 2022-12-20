FROM "mikefarah/yq:4"
RUN [ "mkdir", "/tmp/transfer" ]

ADD resources/* /tmp/transfer/
ADD execute-pipeline.sh execute-pipeline.sh

CMD [ "sh", "-c", "./execute-pipeline.sh" ]
ENTRYPOINT []
