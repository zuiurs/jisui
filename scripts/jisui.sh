#!/bin/bash

iPadAir_Height=1536

for i in "$@"; do
	yes | jisui prepare "$i"
	count=$(ls -U1 "$i" | wc -l)
	jisui comic -v -h ${iPadAir_Height} -pack -skip 1,2,$((${count}-4))-${count} "$i"
done
