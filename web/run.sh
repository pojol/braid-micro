#! /bin/bash

docker run -d -p 8888:8888/tcp braidgo/sankey:latest -consul http://172.17.0.1:8900 -redis redis://172.17.0.1:6379/0
