
Current=d
CurrentDir=/docker/$Current
CurrentFile=$CurrentDir/out.txt
echo -e "_   _   _   _   _   _   _   _   _   _   _   _   start $Current  _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile

Prev01=b
PrevDir01=/docker/$Prev01
PrevFile01=$PrevDir01/out.txt
Prev02=c
PrevDir02=/docker/$Prev02
PrevFile02=$PrevDir02/out.txt
echo $PrevFile01
echo $PrevFile02


cat $PrevFile01 >>  $CurrentFile
cat $PrevFile02 >>  $CurrentFile
cp -rf $PrevDir01 $CurrentDir
cp -rf $PrevDir02 $CurrentDir

echo "created by Current: "$Current  | tee -a $CurrentFile

cd $CurrentDir
mkdir -p Deep01_x/Deep02_x; date >   Deep01_x/Deep02_x/date.txt
mkdir -p Deep01_x/Deep02_y; date >   Deep01_x/Deep02_y/date.txt
mkdir -p Deep01_y/Deep02_x; date >   Deep01_y/Deep02_x/date.txt
mkdir -p Deep01_y/Deep02_y; date >   Deep01_y/Deep02_y/date.txt

# tree
echo $Current  >> /test_dir/date.txt
date >> /test_dir/date.txt

ls -alh /test_dir

sleep 1

echo -e "_   _   _   _   _   _   _   _   _   _   _   _    end $Current   _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile
