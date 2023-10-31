#echo " = = = = = =  "

NAME=c
INPUT=/docker/a/out.txt
OUTPUT=/docker/c/out.txt

cat $INPUT >>  $OUTPUT
echo "created by NAME: "$NAME >> $OUTPUT


#num_args=$#
#i=1
#while [ $i -le $num_args ]; do
#  arg=$(eval echo "\$$i")
#  echo "参数 $i: $arg"
#  i=$((i + 1))
#done
