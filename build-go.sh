#!/bin/bash
find functions/* -type d | while read dir; do
	echo "### Building $dir"
	cd $dir
	GOOS=linux go build -o main
	zip main.zip main
	cd ../..
done