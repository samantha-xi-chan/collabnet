
Prev01=b
PrevDir01=/docker/$PrevDir
PrevFile01=$PrevDir/out.txt
Prev02=c
PrevDir02=/docker/$PrevDir
PrevFile02=$PrevDir/out.txt

Current=d
CurrentDir=/docker/$Current
CurrentFile=$CurrentDir/out.txt

cat $PrevFile01 >>  $CurrentFile
cat $PrevFile02 >>  $CurrentFile
cp -rf $PrevDir01 $CurrentDir
cp -rf $PrevDir02 $CurrentDir

echo "created by Current: "$Current  | tee -a $CurrentFile

echo -e "_   _   _   _   _   _   _   _   _   _   _   _   start $Current  _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile
cd $CurrentDir
mkdir -p Deep01_x/Deep02_x; date >   Deep01_x/Deep02_x/date.txt
mkdir -p Deep01_x/Deep02_y; date >   Deep01_x/Deep02_y/date.txt
mkdir -p Deep01_y/Deep02_x; date >   Deep01_y/Deep02_x/date.txt
mkdir -p Deep01_y/Deep02_y; date >   Deep01_y/Deep02_y/date.txt
date >   date.txt

tree
free -h

echo -e "_   _   _   _   _   _   _   _   _   _   _   _    end $Current   _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile
