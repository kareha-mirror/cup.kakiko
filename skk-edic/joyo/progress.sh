#!/bin/sh

done=$(
  awk '
    FNR == 1 { stop = 0 }

    /^\;----$/ { stop = 1 }

    !stop && /%%%/ { n++ }

    END { print n }
  ' *.txt
)
total=2136

percent=$(echo "scale=2; $done * 100 / $total" | bc)

printf "Progress: %d / %d (%.2f%%)\n" "$done" "$total" "$percent"
