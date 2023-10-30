#echo " = = = = = =  "

NAME=a
OUTPUT=/docker/a/out.txt

echo "created by NAME: "$NAME  | tee -a $OUTPUT

#num_args=$#
#i=1
#while [ $i -le $num_args ]; do
#  arg=$(eval echo "\$$i")
#  echo "参数 $i: $arg"
#  i=$((i + 1))
#done

echo "a going to sleep 1" | tee -a $OUTPUT
sleep 1
echo "a going to sleep a 1" | tee -a $OUTPUT
sleep 1
echo "a going to sleep b 1" | tee -a $OUTPUT
sleep 1
echo "a end of sleep 1" | tee -a $OUTPUT
#exit 2
