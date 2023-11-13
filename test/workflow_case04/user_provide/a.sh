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

loop=2
i=1
while [ $i -le $loop ]; do
  echo "loop $i: " | tee -a $OUTPUT
  TimeSecond=3
  echo "a going to sleep: " $TimeSecond | tee -a $OUTPUT
  sleep $TimeSecond
  i=$((i + 1))
done
