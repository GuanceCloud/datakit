# copy all sql to docker container
find . -maxdepth 1 -name "*.sql" -print0 | xargs -0 -I {} docker cp {} $1:/
