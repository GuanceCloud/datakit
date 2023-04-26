printf ">>>>>>>>>>>>>>>>>>>>>>>>>>\n"
printf ">>> total code\n"
printf ">>>>>>>>>>>>>>>>>>>>>>>>>>\n"
cloc . --exclude-dir=vendor

printf ">>>>>>>>>>>>>>>>>>>>>>>>>>\n"
printf ">>> testing code\n"
printf ">>>>>>>>>>>>>>>>>>>>>>>>>>\n"
find . -name '*.go' | grep test | xargs cloc
