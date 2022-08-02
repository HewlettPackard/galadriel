# ubuntu:20.04 (linux/amd64)
FROM ubuntu@sha256:b2339eee806d44d6a8adc0a790f824fb71f03366dd754d400316ae5a7e3ece3e

ENV TZ=US

RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone \
&& apt-get update -y && apt-get install wget curl git golang vim sudo -y \
&& rm -rf /var/lib/apt/lists/* \
&& useradd -rm -d /home/ubuntu -s /bin/bash -g root -G sudo -u 1001 -p "$(openssl passwd -6 ubuntu)" ubuntu \
&& useradd -rm -d /home/cpqd -s /bin/bash -g root -G sudo -u 1002 -p "$(openssl passwd -6 cpqd)" cpqd\
&& useradd -rm -d /home/spire -s /bin/bash -g root -G sudo -u 1003 -p "$(openssl passwd -6 spire)" spire \
&& curl -s -N -L https://github.com/spiffe/spire/releases/download/v1.3.3/spire-1.3.3-linux-x86_64-glibc.tar.gz | tar xz  \
&& mv /spire-1.3.3 /spire \
&& cp /spire/bin/spire-server /usr/bin/spire-server \
&& cp /spire/bin/spire-agent /usr/bin/spire-agent

ADD dev/spire-server/conf/server/ /spire/conf/server/
ADD dev/spire-server/script-server.sh spire/

RUN chmod +x /spire/script-server.sh

WORKDIR /spire

ENTRYPOINT ./script-server.sh ; /bin/bash