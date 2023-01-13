printf ">>>>>>>>>>>>>>>>>>>>>>>>>>\n"
printf ">>> total code\n"
printf ">>>>>>>>>>>>>>>>>>>>>>>>>>\n"
cloc . --exclude-dir=gen

printf ">>>>>>>>>>>>>>>>>>>>>>>>>>\n"
printf ">>> testing code\n"
printf ">>>>>>>>>>>>>>>>>>>>>>>>>>\n"
find . -name '*.go' | grep test | xargs cloc
