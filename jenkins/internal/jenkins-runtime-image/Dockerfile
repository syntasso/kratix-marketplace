FROM jenkins/jenkins:2.504.1-jdk17

LABEL org.opencontainers.image.authors="kratix@syntasso.io"
LABEL org.opencontainers.image.source=https://github.com/syntasso/kratix-marketplace

ARG JENKINS_UC
ARG JENKINS_UC_DOWNLOAD
ARG JENKINS_PLUGIN_INFO
ARG JENKINS_INCREMENTALS_REPO_MIRROR

COPY --chown=jenkins:jenkins plugins.txt /usr/share/jenkins/ref/plugins.txt

RUN set -eux; \
    if [ -n "${JENKINS_UC:-}" ]; then export JENKINS_UC; fi; \
    if [ -n "${JENKINS_UC_DOWNLOAD:-}" ]; then export JENKINS_UC_DOWNLOAD; fi; \
    if [ -n "${JENKINS_PLUGIN_INFO:-}" ]; then export JENKINS_PLUGIN_INFO; fi; \
    if [ -n "${JENKINS_INCREMENTALS_REPO_MIRROR:-}" ]; then export JENKINS_INCREMENTALS_REPO_MIRROR; fi; \
    jenkins-plugin-cli --latest false --plugin-file /usr/share/jenkins/ref/plugins.txt
