#!/bin/sh

done=$(grep -h '%%%' *.txt | wc -l)
total=2136

percent=$(echo "scale=2; $done * 100 / $total" | bc)

printf "Progress: %d / %d (%.2f%%)\n" "$done" "$total" "$percent"
