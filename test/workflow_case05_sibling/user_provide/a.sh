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

TimeSecond=40

echo "a going to sleep: " $TimeSecond | tee -a $OUTPUT
sleep $TimeSecond
echo "end of  sleep :" $TimeSecond | tee -a $OUTPUT
