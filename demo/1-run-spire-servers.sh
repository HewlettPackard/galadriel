#!/bin/bash
set -e
source util.sh

one ./bin/spire-server run -config one.org/spire/conf/server/server.conf
two ./bin/spire-server run -config two.org/spire/conf/server/server.conf

wait
