#!/bin/bash

# Install waque:
#		$ npm i -g waque
#		$ waque login

rm -rf .docs
mkdir -p .docs
cp man/summary.md .docs/
dist/datakit-darwin-amd64/datakit --ignore-manuals demo --export-manuals .docs
waque upload .docs/*.md
