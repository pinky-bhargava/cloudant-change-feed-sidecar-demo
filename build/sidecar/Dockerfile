FROM registry.access.redhat.com/ubi8/ubi:8.3
LABEL maintainer="pinky.bhargava@in.ibm.com"
ARG SKIP_LINT=false
ENV SERVICE=sidecar
ENV GOPATH=/usr
RUN mkdir -p $GOPATH/src/github.com/cloudant-change-feed-sidecar-demo
WORKDIR $GOPATH/src/github.com/cloudant-change-feed-sidecar-demo
COPY . .
RUN export GOPROXY=https://proxy.golang.org/ && \
    chmod +x ./scripts/setup_sidecar.sh && ./scripts/setup_sidecar.sh
CMD ["sidecar"]
