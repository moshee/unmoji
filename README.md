Command-line interface to [mojibake](http://github.com/moshee/mojibake).

```
Usage of unmoji:
  -args=false: Decode arguments instead of from STDIN
  -encs="utf-8": Comma-separated encoding path
  -f=false: Skip errors
  -really=false: When -rename is given, actually do the renaming instead of just showing what will happen
  -rename=false: Like args but rename the named files to the decoded values
Available encoding options for -encs:
    utf-8, utf8, cp473: CP473 (assume UTF-8 input misinterpreted as ASCII)
    sjis, shift-jis, cp932: CP932 (assume Shift-JIS input misinterpreted as UTF-8)
    cjk, cp936: CP936 (assume CJK input misinterpreted as UTF-8)
```
