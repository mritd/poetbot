FROM golang:1.15-alpine3.12 AS builder

ENV SRC_PATH ${GOPATH}/src/github.com/mritd/poetbot

COPY . ${SRC_PATH}

WORKDIR ${SRC_PATH}

RUN set -ex \
    && apk add --update alpine-sdk linux-headers git zlib-dev openssl-dev gperf php cmake \
    && git clone https://github.com/tdlib/td.git \
    && (cd td && git checkout v1.7.0 \
    && rm -rf build && mkdir build && cd build \
    && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/usr/local .. \
    && cmake --build . --target install -- -j$(nproc) \
    && cd ../../ && ls -l /usr/local) \
    && export BUILD_VERSION=$(cat version) \
    && export BUILD_DATE=$(date "+%F %T") \
    && export COMMIT_SHA1=$(git rev-parse HEAD) \
    && go install -ldflags \
        "-X 'main.version=${BUILD_VERSION}' \
        -X 'main.buildDate=${BUILD_DATE}' \
        -X 'main.commitID=${COMMIT_SHA1}'" \
    && scanelf --needed --nobanner /go/bin/poetbot | \
        awk '{ gsub(/,/, "\nso:", $2); print "so:" $2 }' | \
        sort -u | tee /dep_so


FROM alpine:3.12

ARG TZ="Asia/Shanghai"

ENV TZ ${TZ}
ENV LANG en_US.UTF-8
ENV LC_ALL en_US.UTF-8
ENV LANGUAGE en_US:en

# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

COPY --from=builder /dep_so /dep_so
COPY --from=builder /go/bin/poetbot /usr/bin/poetbot
COPY --from=builder /go/src/github.com/mritd/poetbot/poet.txt /poet.txt

RUN set -ex \
    && apk add bash tzdata ca-certificates \
    && apk add $(cat /dep_so) \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && rm -rf /var/cache/apk/* /dep_so

VOLUME /data

CMD ["poetbot"]
