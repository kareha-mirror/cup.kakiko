all: skk-edic-joyo.txt

skk-edic-joyo.txt: joyo/*.txt
	../skk-edic-merge joyo/*.txt > skk-edic-joyo.txt
	cp skk-edic-joyo.txt ../cmd/kakiko/

clean:
	rm -f skk-edic-joyo.txt
	rm -f ../cmd/kakiko/skk-edic-joyo.txt
