#!/bin/bash
set -e
source util.sh

spire_one_socket=/tmp/one.org/spire-server/private/api.sock
spire_two_socket=/tmp/two.org/spire-server/private/api.sock

# Get current CA from SPIRE Server
./bin/spire-server bundle show -socketPath ${spire_one_socket} > ./one.org/spire/conf/agent/root_ca.crt
./bin/spire-server bundle show -socketPath ${spire_two_socket} > ./two.org/spire/conf/agent/root_ca.crt

# Get join tokens
token_one=$(./bin/spire-server token generate -socketPath ${spire_one_socket} -spiffeID spiffe://one.org/my-agent | grep Token | cut -c 8-100)
token_two=$(./bin/spire-server token generate -socketPath ${spire_two_socket} -spiffeID spiffe://two.org/my-agent | grep Token | cut -c 8-100)

# Start agents
one ./bin/spire-agent run -config ./one.org/spire/conf/agent/agent.conf -joinToken ${token_one}
two ./bin/spire-agent run -config ./two.org/spire/conf/agent/agent.conf -joinToken ${token_two}

wait
