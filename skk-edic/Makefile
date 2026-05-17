all: skk-edic-default.txt

skk-edic-default.txt: skk-edic-joyo.txt skk-edic-handcopy.txt
	../skk-edic-merge skk-edic-joyo.txt skk-edic-handcopy.txt > skk-edic-default.txt
	cp skk-edic-default.txt ../cmd/kakiko/

skk-edic-joyo.txt: joyo/*.txt
	../skk-edic-merge joyo/*.txt > skk-edic-joyo.txt
	cp skk-edic-joyo.txt ../cmd/kakiko/

skk-edic-handcopy.txt: handcopy/station/config/skk-edic-diff.txt handcopy/computer/config/skk-edic-diff.txt
	../skk-edic-merge handcopy/station/config/skk-edic-diff.txt handcopy/computer/config/skk-edic-diff.txt > skk-edic-handcopy.txt

clean:
	rm -f skk-edic-joyo.txt
	rm -f skk-edic-handcopy.txt
	rm -f skk-edic-default.txt
	rm -f ../cmd/kakiko/skk-edic-joyo.txt
	rm -f ../cmd/kakiko/skk-edic-default.txt
