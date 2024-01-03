
Current=c
CurrentDir=/docker/$Current
CurrentFile=$CurrentDir/out.txt
echo $CurrentFile
echo -e "_   _   _   _   _   _   _   _   _   _   _   _   start $Current  _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile


sleep 1

Prev=a
PrevDir=/docker/$Prev
PrevFile=$PrevDir/out.txt

cat $PrevFile >>  $CurrentFile
cp -rf $PrevDir $CurrentDir
echo "created by Current: "$Current  | tee -a $CurrentFile

cd $CurrentDir
mkdir -p Deep01_x/Deep02_x; date >   Deep01_x/Deep02_x/date.txt
mkdir -p Deep01_x/Deep02_y; date >   Deep01_x/Deep02_y/date.txt
mkdir -p Deep01_y/Deep02_x; date >   Deep01_y/Deep02_x/date.txt
mkdir -p Deep01_y/Deep02_y; date >   Deep01_y/Deep02_y/date.txt

# tree
echo $Current  >> /test_dir/date.txt
date >> /test_dir/date.txt

ls /docker/c

timeout=1
echo  $Current "going to sleep " $timeout " seconds"
sleep $timeout
echo  $Current "end of sleep " $timeout " seconds"

echo -e "_   _   _   _   _   _   _   _   _   _   _   _    end $Current   _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile
