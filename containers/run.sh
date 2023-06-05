#!/bin/bash

docker run -d --name galadrieldb -p 5432:5432 galadrieldb
docker run -d --name pgadmin -p 82:80 pgadmin

