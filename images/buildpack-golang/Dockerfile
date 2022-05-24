# Basic golang buildpack

FROM golang:1.18.2

# Commit details

ARG commit
ENV IMAGE_COMMIT=$commit
LABEL io.kyma-project.test-infra.commit=$commit

# Versions

ENV ARCH amd64
ENV KUBEBUILDER_VERSION 2.3.2
ENV KUSTOMIZE_VERSION 3.8.4

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# apt and install additional packages
RUN apt-get update && apt-get install -y --no-install-recommends \
        rsync \
        procps \
        pkg-config \
        libgit2-dev \
        && apt-get clean \
        && rm -rf /var/lib/apt/lists/*

# Install kubebuilder
ENV PATH /usr/local/kubebuilder/bin:$PATH
ENV PATH /home/prow/go/bin:$PATH

RUN curl -fLSs -o kubebuilder.tar.gz "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kubebuilder.tar.gz && \
    rm kubebuilder.tar.gz && \
    mv "kubebuilder_${KUBEBUILDER_VERSION}_linux_${ARCH}" kubebuilder && \
    mv kubebuilder /usr/local/ && \
    curl -fLSs -o kustomize.tar.gz "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v${KUSTOMIZE_VERSION}/kustomize_v${KUSTOMIZE_VERSION}_linux_${ARCH}.tar.gz" && \
    tar -zxvf kustomize.tar.gz && \
    rm kustomize.tar.gz && \
    mv kustomize /usr/local/bin/kustomize

COPY ./license-puller.sh /license-puller.sh
ENV LICENSE_PULLER_PATH=/license-puller.sh
