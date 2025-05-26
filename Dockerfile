# Build the manager binary
FROM container-registry.oracle.com/os/oraclelinux:8 as builder
ENV GOLANG_VERSION=1.23.3
ARG GOLANG_VERSION
RUN if [ -n "$GOLANG_VERSION" ]; then \
        curl -LJO https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz &&\
        rm -rf /usr/local/go && tar -C /usr/local -xzf go${GOLANG_VERSION}.linux-amd64.tar.gz &&\
        rm go${GOLANG_VERSION}.linux-amd64.tar.gz; \
    fi
ENV PATH=${GOLANG_VERSION:+"${PATH}:/usr/local/go/bin"}

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY internal/ internal/
COPY common/ common/
COPY ol8_oracle_instantclient21.repo /etc/yum.repos.d/
RUN yum-config-manager --enable ol8_oracle_instantclient21
RUN yum-config-manager --enable ol8_codeready_builder
RUN yum install -y libtool gcc gdb libgcc glibc glibc-devel \
    glibc-common  gcc-c++  glibc-static
RUN yum install -y oracle-instantclient-release-el8 \
    oracle-instantclient-basic oracle-instantclient-devel \
    net-tools oracle-instantclient-sqlplus make

# Build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM container-registry.oracle.com/os/oraclelinux:8
COPY ol8_oracle_instantclient21.repo /etc/yum.repos.d/
RUN yum-config-manager --enable ol8_oracle_instantclient21
RUN yum-config-manager --enable ol8_codeready_builder
RUN yum install -y libtool gcc gdb libgcc glibc glibc-devel \
    glibc-common  gcc-c++  glibc-static
RUN yum install -y oracle-instantclient-release-el8 \
    oracle-instantclient-basic oracle-instantclient-devel \
    net-tools oracle-instantclient-sqlplus

#WORKDIR /
#COPY --from=builder /workspace/manager .
#RUN useradd -u 1002 nonroot
#USER nonroot

WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532


ENTRYPOINT ["/manager"]
