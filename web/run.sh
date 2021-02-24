#! /bin/bash

docker run -d -p 8888:8888/tcp braidgo/web:latest -consul http://172.17.0.1:8500 -redis redis://172.17.0.1:6379/0
