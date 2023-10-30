#echo " = = = = = =  "

NAME=b
INPUT=/docker/a/out.txt
OUTPUT=/docker/b/out.txt

cat $INPUT >>  $OUTPUT
echo "created by NAME: "$NAME >> $OUTPUT


#num_args=$#
#i=1
#while [ $i -le $num_args ]; do
#  arg=$(eval echo "\$$i")
#  echo "参数 $i: $arg"
#  i=$((i + 1))
#done


echo "b going to sleep 1" | tee -a $OUTPUT
sleep 1
echo "a end of sleep " | tee -a $OUTPUT
