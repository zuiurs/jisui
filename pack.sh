#!/bin/bash

#-----------
EXT=jpg
#-----------

work_dir="$(/usr/bin/dirname $0)/$1"

cd "$work_dir"

files=($(/bin/ls *.$EXT))
num=($(/usr/bin/seq -w 001 ${#files[@]}))

# e.g.) MAR_1_001.jpg -> MAR_1
prefix=${files[0]%_*}

for (( i=0; i<${#files[@]}; i++ )); do
	/bin/echo ${files[$i]} "->" ${prefix}_${num[$i]}.${EXT}
done


