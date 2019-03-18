FROM centos:7 as builder
LABEL maintainer "Devtools <devtools@redhat.com>"
LABEL author "Konrad Kleine <kkleine@redhat.com>"
ENV LANG=en_US.utf8
ENV GOPATH /tmp/go
ARG GO_PACKAGE_PATH=github.com/redhat-developer/app-service
ARG VERBOSE

RUN yum install epel-release -y \
  && yum install --enablerepo=centosplus install -y --quiet \
    findutils \
    git \
    golang \
    make \
    procps-ng \
    tar \
    wget \
    which \
    bc \
  && yum clean all \
  && mkdir -p $GOPATH/bin && chmod a+rwx $GOPATH \
  && curl -L -s https://github.com/golang/dep/releases/download/v0.5.1/dep-linux-amd64 -o dep \
  && echo "7479cca72da0596bb3c23094d363ea32b7336daa5473fa785a2099be28ecd0e3  dep" > dep-linux-amd64.sha256 \
  && sha256sum -c dep-linux-amd64.sha256 \
  && rm dep-linux-amd64.sha256 \
  && chmod +x ./dep \
  && mv dep $GOPATH/bin/dep
ENV PATH=$PATH:$GOPATH/bin

COPY . ${GOPATH}/src/${GO_PACKAGE_PATH}

WORKDIR ${GOPATH}/src/${GO_PACKAGE_PATH}

RUN make VERBOSE=${VERBOSE}
RUN make VERBOSE=${VERBOSE} test

#--------------------------------------------------------------------

FROM centos:7
LABEL maintainer "Devtools <devtools@redhat.com>"
LABEL author "Konrad Kleine <kkleine@redhat.com>"
ENV LANG=en_US.utf8
ENV APP_INSTALL_PREFIX=/usr/local/app-server

ENV GOPATH=/tmp/go
ARG GO_PACKAGE_PATH=github.com/redhat-developer/app-service

# Create a non-root user and a group with the same name: "appserver"
ENV APP_USER_NAME=appserver
RUN useradd --no-create-home -s /bin/bash ${APP_USER_NAME}

COPY --from=builder ${GOPATH}/src/${GO_PACKAGE_PATH}/out/app-server ${APP_INSTALL_PREFIX}/bin/app-server

# From here onwards, any RUN, CMD, or ENTRYPOINT will be run under the following user
USER ${APP_USER_NAME}

WORKDIR ${APP_INSTALL_PREFIX}
ENTRYPOINT [ "./bin/app-server" ]

EXPOSE 8080