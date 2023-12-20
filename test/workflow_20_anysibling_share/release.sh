

rm -rf ./test_dir* ;
make build;
mkdir test_dir ;

cp -r test/case03 test_dir ;
cp -r out/*  test_dir ;
cp -r config test_dir ;

tar -czvf release/"test_dir_$(date +'%Y%m%d.%H%M%S').tar.gz" test_dir

rm -rf test_dir
