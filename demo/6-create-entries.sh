#!/bin/bash
set -e

spire_one_socket=/tmp/one.org/spire-server/private/api.sock
spire_two_socket=/tmp/two.org/spire-server/private/api.sock

# Create entry for greeter-server in one.org
./bin/spire-server entry create \
    -socketPath ${spire_one_socket} \
    -spiffeID spiffe://one.org/greeter-server \
    -parentID spiffe://one.org/my-agent \
    -selector unix:uid:$(id -u) \
    -federatesWith two.org

# Create entry for greeter-client in two.org
./bin/spire-server entry create \
    -socketPath ${spire_two_socket} \
    -spiffeID spiffe://two.org/greeter-client \
    -parentID spiffe://two.org/my-agent \
    -selector unix:uid:$(id -u) \
    -federatesWith one.org
