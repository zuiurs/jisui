#!/bin/bash

#-----------
EXT=jpg
#-----------

while getopts "d" OPT
do
	case $OPT in
		d)	# Dry Run
			dryrun=true
			;;
	esac
done
shift $((OPTIND - 1))

cd "$1"

files=($(/bin/ls *.$EXT))
num=($(/usr/bin/seq -w 001 ${#files[@]}))

# e.g.) MAR_1_001.jpg -> MAR_1
prefix=${files[0]%_*}

for (( i=0; i<${#files[@]}; i++ )); do
	if [[ $dryrun == "true" ]]; then
		/bin/echo ${files[$i]} ${prefix}_${num[$i]}.${EXT}
	else
		/bin/mv ${files[$i]} ${prefix}_${num[$i]}.${EXT}
	fi
done

echo packed!
