#!/bin/bash
set -e
source util.sh

token_one=$(./bin/galadriel-server generate token -t one.org | cut -c 15-100)
token_two=$(./bin/galadriel-server generate token -t two.org | cut -c 15-100)

one ./bin/galadriel-harvester run --config ./one.org/harvester/harvester.conf --token ${token_one}
two ./bin/galadriel-harvester run --config ./two.org/harvester/harvester.conf --token ${token_two}

wait
