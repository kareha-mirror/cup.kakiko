all: skk-edic-handcopy.txt

skk-edic-handcopy.txt: station/config/skk-edic-diff.txt computer/config/skk-edic-diff.txt diary/config/skk-edic-diff.txt science/config/skk-edic-diff.txt world/config/skk-edic-diff.txt philosophy/config/skk-edic-diff.txt
	../../skk-edic-merge station/config/skk-edic-diff.txt computer/config/skk-edic-diff.txt diary/config/skk-edic-diff.txt science/config/skk-edic-diff.txt world/config/skk-edic-diff.txt philosophy/config/skk-edic-diff.txt > skk-edic-handcopy.txt

clean:
	rm -f skk-edic-handcopy.txt
