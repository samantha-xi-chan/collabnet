
Current=a
CurrentDir=/docker/$Current
CurrentFile=$CurrentDir/out.txt

echo " 😄 created by Current: "$Current  | tee -a $CurrentFile

echo -e "_   _   _   _   _   _   _   _   _   _   _   _   start $Current  _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile
cd $CurrentDir
mkdir -p Deep01_x/Deep02_x; date >   Deep01_x/Deep02_x/date.txt
mkdir -p Deep01_x/Deep02_y; date >   Deep01_x/Deep02_y/date.txt
mkdir -p Deep01_y/Deep02_x; date >   Deep01_y/Deep02_x/date.txt
mkdir -p Deep01_y/Deep02_y; date >   Deep01_y/Deep02_y/date.txt
ln -s Deep01_y/Deep02_y/date.txt date.txt
ln -s Deep01_y/Deep02_y/not_exists.txt date_fail.txt

free -h
tree

echo -e "_   _   _   _   _   _   _   _   _   _   _   _    end $Current   _   _   _   _   _   _   _   _   _   _   _   _ " | tee -a $CurrentFile
