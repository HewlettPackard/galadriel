#!/bin/bash
spire-server run -socketPath /tmp/spire-server/hpe/api.sock -config conf/server/hpe.conf &
spire-server run -socketPath /tmp/spire-server/cpqd/api.sock -config conf/server/cpqd.conf &