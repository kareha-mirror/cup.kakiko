# kakikoのステイタス行の色の設定

kakikoのステイタス行の色は、前景色と背景色をそれぞれ、設定ファイルで指定できる。
設定ファイルは通常ホームディレクトリの下の `.config/kakiko/fep.yaml` にある。
その中の `fg-color` で前景色を、 `bg-color` で背景色を指定する。
指定するのはダブルクォートで囲まれた文字列である。

端末のデフォルトの色にするには `"default"` を指定する。

16色のどれかを指定するには次のどれかを指定する。

```
"black" "red" "green" "yellow" "blue" "magenta" "cyan" "white"
"bright black" "bright red" "bright green" "bright yellow"
"bright blue" "bright magenta" "bright cyan" "bright white"
```

256色のどれかを指定するには `"0"` から `"255"` の間の数値を文字列で指定する。

RGBのTrueColorを指定するには `"RRGGBB"` の形式で16進数で値を指定する。
例えば `"0080ff"` など。

デフォルトでは前景色に `"252"`、背景色に `"235"` が指定されている。
