all: skk-edic-joyo.txt

skk-edic-joyo.txt: joyokanjihyo_20101130/*.txt
	../skk-edic-merge joyokanjihyo_20101130/*.txt > skk-edic-joyo.txt

clean:
	rm -f skk-edic-joyo.txt
