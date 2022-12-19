FROM "ruby"
RUN [ "mkdir", "/tmp/transfer" ]

ADD resources/redis-instance.yaml /tmp/transfer/redis-instance.yaml
ADD execute-pipeline.rb execute-pipeline.rb

CMD [ "ruby", "./execute-pipeline.rb"]
ENTRYPOINT []
