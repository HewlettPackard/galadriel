#!/bin/bash
set -e

./bin/galadriel-server create member --trustDomain one.org
./bin/galadriel-server create member --trustDomain two.org
./bin/galadriel-server create relationship -a one.org -b two.org
