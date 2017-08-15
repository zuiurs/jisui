#!/bin/bash

#-----------
EXT=jpg
#-----------

dest=$(basename "$1").comic
cp -ai "$1" "${dest}"

cd "${dest}"

echo GrayScalling...
mogrify -monitor -channel R -separate *.${EXT}
echo Whitening...
mogrify -monitor -level 0%,90% *.${EXT}
echo Adjust Black Level...
mogrify -monitor -level 35%,100% *.${EXT}
echo Complete!



