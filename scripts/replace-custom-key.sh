#!/bin/bash
# Replace exist words in markdown docs
find . -type f -name '*.md' -exec sed -i '' s/观测云/"<<<custom_key.brand_name>>>"/g {} +
find . -type f -name '*.md' -exec sed -i '' s/"Guance Cloud"/"<<<custom_key.brand_name>>>"/g {} +
find . -type f -name '*.md' -exec sed -i '' s/"guance.com"/"<<<custom_key.static_domain>>>"/g {} +
