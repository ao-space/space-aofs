# Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM alpine:edge as builder

WORKDIR /opt/

COPY . .

RUN apk add wget \
        && cd / && arch=$(uname -m) \
        && case "$arch" in x86_64) arch="amd64"; ;; \
        aarch64) arch="arm64"; ;; \
        riscv64) arch="riscv64"; ;; \
        *) arch="Unknown"; ;; esac \
        && wget https://go.dev/dl/go1.21.1.linux-$arch.tar.gz \
        && tar -xzvf go1.21.1.linux-$arch.tar.gz \
        && export PATH=$PATH:/go/bin && cd /opt/ \
        && go env -w CGO_ENABLED=0 \
        && go env -w GO111MODULE=on \
        && go build -o space-aofs

FROM alpine:edge

RUN apk --no-cache add wget ntfs-3g \
        lsblk exfat-utils fuse-exfat usbutils mediainfo eudev
COPY --from=builder /opt/space-aofs /usr/bin
COPY --from=builder /opt/template/ /tmp/
RUN chmod +x /usr/bin/space-aofs
HEALTHCHECK --interval=30s --timeout=15s \
    CMD wget --output-document=/dev/null http://localhost:2001/space/v1/api/status?userId=1

ENTRYPOINT ["/usr/bin/space-aofs"]
