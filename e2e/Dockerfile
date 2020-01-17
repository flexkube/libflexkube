FROM hashicorp/terraform:light

RUN apk add -U make bash bash-completion vim

RUN wget https://get.helm.sh/helm-v3.0.1-linux-amd64.tar.gz -O - | tar -xzvf - linux-amd64/helm && \
    mv linux-amd64/helm /usr/local/bin/

RUN wget -O /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/$(wget -q -O- https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
    && chmod +x /usr/local/bin/kubectl \
    && echo -e 'source /usr/share/bash-completion/bash_completion\nsource <(kubectl completion bash)' >> $HOME/.bashrc

RUN echo 'plugin_cache_dir = "/root/.terraform.d.host/plugin-cache"' >> /root/.terraformrc

RUN helm repo add stable https://kubernetes-charts.storage.googleapis.com

RUN helm repo add flexkube https://flexkube.github.io/charts/

ENV KUBECONFIG=/root/libflexkube/e2e/kubeconfig
