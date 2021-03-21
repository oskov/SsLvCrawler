#!/bin/bash

docker-compose up -d

sleep 5

./retailerTool sell # default riga
./retailerTool rent # default riga
./retailerTool sell --city=jelgava
./retailerTool rent --city=jelgava
./retailerTool sell --city=jelgava-and-reg
./retailerTool rent --city=jelgava-and-reg
./retailerTool sell --city=jurmala
./retailerTool rent --city=jurmala
./retailerTool sell --city=riga-region
./retailerTool rent --city=riga-region

docker stop mariadb-server-golang-retail
