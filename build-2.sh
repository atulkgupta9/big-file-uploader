#!/bin/bash

cd server
go build -o ../bigf
cd .. && ./bigf 3002