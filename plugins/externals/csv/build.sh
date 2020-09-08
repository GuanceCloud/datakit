mkdir -p $1/csv/rpc
mkdir -p $1/csv/csvkit

cp plugins/externals/csv/main.py            $1/csv/
cp plugins/externals/csv/requirement.txt    $1/csv/
cp plugins/externals/csv/csvkit/*.py        $1/csv/csvkit/
