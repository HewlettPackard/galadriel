#!/bin/bash
set -e
source util.sh

agent_one_socket=/tmp/one.org/spire-agent/private/api.sock
agent_two_socket=/tmp/two.org/spire-agent/private/api.sock

cleanup() {
    pkill greeter-server greeter-client
}
trap cleanup EXIT

# Run greeter server and client
server ./bin/greeter-server --workloadapi unix://${agent_one_socket}; sleep 5
client ./bin/greeter-client --workloadapi unix://${agent_two_socket}

wait
