# kakiko - A SKK Frontend Processor for Text Terminals

kakiko is a frontend processor for text terminals which enables Japanese text input by using SKK method.

## Build

kakiko is written in Go.
You need Go build tools.
Run `make` and executable file `kakiko` will be built.

## Run

When you run `kakiko` without any argument, the default shell for the user will be invoked which is wrapped by kakiko SKK FEP.

When you run `kakiko` with command argument, the command will be invoked instead of the default shell.

## Dictionary

A minimal dictionary is embedded on kakiko executable file.
If you want to use larger dictionary, download [skk-edic-legacy-l.cdb](https://tea.kareha.org/ja/skk-e/raw/branch/main/legacy-cdb/skk-edic-legacy-l.cdb) and place it `$HOME/.config/kakiko/skk-edic-legacy-l.cdb`.
