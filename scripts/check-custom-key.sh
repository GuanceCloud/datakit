#!/bin/bash

keywords=(
观测云         # -> <<<custom_key.brand_name>>>
"Guance Cloud" # -> <<<custom_key.brand_name>>>
"Guance"       # -> <<<custom_key.brand_name>>>
guance.com     # -> <<<custom_key.brand_main_domain>>>
# add more...
)

ON_ERROR=0
doc_dir=$1
rm -rf cck.out
for word in "${keywords[@]}"; do
	echo "checking '$word' under $doc_dir..."
	tmp_file=$(mktemp)

	grep -nrw "$word" $doc_dir | tee -a cck.out
done

while read -r line; do
	echo "ERROR: found '$word' at $line"
	ON_ERROR=1
done < cck.out

if [ "$ON_ERROR" -eq 1 ]; then
	exit 1
else
	echo "keywords checking done."
	exit 0
fi
