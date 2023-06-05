#!/bin/bash

cd ./database
docker build -t galadrieldb .
cd -

cd ./pgadmin
docker build -t pgadmin .
cd -