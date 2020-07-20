mkdir -p $1/csv/rpc
mkdir -p $1/csv/csvkit

cp plugins/external/csv/main.py            $1/csv/
cp plugins/external/csv/requirement.txt    $1/csv/
cp plugins/external/csv/rpc/*.py           $1/csv/rpc/
cp plugins/external/csv/csvkit/*.py        $1/csv/csvkit/
