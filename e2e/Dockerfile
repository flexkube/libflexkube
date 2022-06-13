FROM golang:1.18-alpine3.15

RUN apk add -U make bash bash-completion vim coreutils

RUN wget https://get.helm.sh/helm-v3.8.1-linux-amd64.tar.gz -O - | tar -xzvf - linux-amd64/helm && \
    mv linux-amd64/helm /usr/local/bin/

RUN wget -O /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$(wget -q -O- https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && echo -e 'source /usr/share/bash-completion/bash_completion\nsource <(kubectl completion bash)' >> $HOME/.bashrc

RUN VERSION=0.56.7 OS=linux && \
    wget "https://github.com/vmware-tanzu/sonobuoy/releases/download/v${VERSION}/sonobuoy_${VERSION}_${OS}_amd64.tar.gz" -O sonobuoy.tar.gz && \
    tar -xzf sonobuoy.tar.gz -C /usr/local/bin && \
    chmod +x /usr/local/bin/sonobuoy && \
    rm sonobuoy.tar.gz

RUN helm repo add flexkube https://flexkube.github.io/charts/

ENV KUBECONFIG=/root/libflexkube/e2e/kubeconfig
