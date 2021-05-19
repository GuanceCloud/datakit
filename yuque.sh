#!/bin/bash

# Install waque:
#		$ npm i -g waque
#		$ waque login

rm -rf .docs
mkdir -p .docs
cp man/summary.md .docs/
man_version=`git tag -l | sort -nr | head -n 1` # use latest tag version
dist/datakit-darwin-amd64/datakit --ignore demo --export-manuals .docs --man-version $man_version
waque upload .docs/*.md
