all: skk-edic-joyo.txt skk-edic-overlay.txt

skk-edic-joyo.txt: joyo/*.txt
	../skk-edic-merge joyo/*.txt > skk-edic-joyo.txt
	cp skk-edic-joyo.txt ../cmd/kakiko/

skk-edic-overlay.txt: skk-edic-kihon.txt skk-edic-handcopy.txt
	../skk-edic-merge skk-edic-kihon.txt skk-edic-handcopy.txt > skk-edic-overlay.txt
	cp skk-edic-overlay.txt ../cmd/kakiko/

skk-edic-kihon.txt: kihon/*.txt
	../skk-edic-merge kihon/*.txt > skk-edic-kihon.txt

skk-edic-handcopy.txt: handcopy/station/config/skk-edic-diff.txt handcopy/computer/config/skk-edic-diff.txt handcopy/diary/config/skk-edic-diff.txt
	../skk-edic-merge handcopy/station/config/skk-edic-diff.txt handcopy/computer/config/skk-edic-diff.txt handcopy/diary/config/skk-edic-diff.txt > skk-edic-handcopy.txt

clean:
	rm -f skk-edic-joyo.txt
	rm -f skk-edic-kihon.txt
	rm -f skk-edic-handcopy.txt
	rm -f skk-edic-overlay.txt
	rm -f ../cmd/kakiko/skk-edic-joyo.txt
	rm -f ../cmd/kakiko/skk-edic-overlay.txt
