FROM "alpine"

RUN apk update && apk add --no-cache yq kubectl curl ruby
RUN curl https://releases.hashicorp.com/terraform/1.7.1/terraform_1.7.1_linux_amd64.zip -o terraform.zip && \
    unzip terraform.zip && \
    mv terraform /usr/local/bin/terraform && \
    rm terraform.zip

COPY scripts/* /usr/bin/
ADD resources resources

RUN chmod +x /usr/bin/*

ENTRYPOINT []
