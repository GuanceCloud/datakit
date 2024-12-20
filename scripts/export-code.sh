find . -name '*.go' | grep -vE vendor |xargs cat > code.txt
