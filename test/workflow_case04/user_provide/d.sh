#echo " = = = = = =  "

NAME=d
INPUT_B=/docker/b/out.txt
INPUT_C=/docker/c/out.txt
OUTPUT=/docker/d/out.txt

cat $INPUT_B >>  $OUTPUT
cat $INPUT_C >>  $OUTPUT
echo "created by NAME: "$NAME >> $OUTPUT

#num_args=$#
#i=1
#while [ $i -le $num_args ]; do
#  arg=$(eval echo "\$$i")
#  echo "参数 $i: $arg"
#  i=$((i + 1))
#done

sleep  1



