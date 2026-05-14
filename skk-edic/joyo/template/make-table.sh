#!/bin/sh
cat table-orig.txt | grep '^ ' | sed 's/^             //' | sed 's/^            //' | sed 's/ \+/,/g' | sed 's/（/(/g' | sed 's/）/)/g' | sed 's/，/ /g ' | sed 's/⇔/<->/g' > table.txt
