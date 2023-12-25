
Prev=a
PrevDir=/docker/$PrevDir
PrevFile=$PrevDir/out.txt

Current=b
CurrentDir=/docker/$Current
CurrentFile=$CurrentDir/out.txt

cat $PrevFile >>  $CurrentFile
cp -rf $PrevDir $CurrentDir
echo "created by Current: "$Current  | tee -a $CurrentFile

echo -e "_   _   _   _   _   _   _   _   _   _   _   _   start $Current  _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile
cd $CurrentDir
mkdir -p Deep01_x/Deep02_x; date >   Deep01_x/Deep02_x/date.txt
mkdir -p Deep01_x/Deep02_y; date >   Deep01_x/Deep02_y/date.txt
mkdir -p Deep01_y/Deep02_x; date >   Deep01_y/Deep02_x/date.txt
mkdir -p Deep01_y/Deep02_y; date >   Deep01_y/Deep02_y/date.txt
date >   date.txt

tree

echo -e "_   _   _   _   _   _   _   _   _   _   _   _    end $Current   _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile
