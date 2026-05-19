all: skk-edic-joyo.txt skk-edic-overlay.txt

skk-edic-joyo.txt: joyo/*.txt
	../skk-edic-merge joyo/*.txt > skk-edic-joyo.txt
	cp skk-edic-joyo.txt ../cmd/kakiko/

skk-edic-overlay.txt: skk-edic-kihon.txt
	make -C handcopy
	../skk-edic-merge skk-edic-kihon.txt handcopy/skk-edic-handcopy.txt > skk-edic-overlay.txt
	cp skk-edic-overlay.txt ../cmd/kakiko/

skk-edic-kihon.txt: kihon/*.txt
	../skk-edic-merge kihon/*.txt > skk-edic-kihon.txt

clean:
	make -C handcopy clean
	rm -f skk-edic-joyo.txt
	rm -f skk-edic-kihon.txt
	rm -f skk-edic-overlay.txt
	rm -f ../cmd/kakiko/skk-edic-joyo.txt
	rm -f ../cmd/kakiko/skk-edic-overlay.txt
