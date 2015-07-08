Lycos
===
GAE/Goで動く長期人狼(開発中)  
# Introduction
Lycos is "Are you a Werewolf?(Jinrou)" game on Google App Engine.  
Supported in Japanese.  
This game often played in short(2-3 hours) or long(5-7 days) term.  
Those version are very different, Lycos is supporting Long-term Game.  

This software is alpha version, So some functions are not implemented.  
"Lycos" means "Werewolf" in Greek.  

# Requirements
+ Golang
+ GAE/Go SDK

# Library
+ Goon
+ go-yaml

# Usage
Setup Google App Engine SDK for Go  
https://cloud.google.com/appengine/downloads  

### Test on Local
cp app_sample.yaml app.yaml  
And
set own unique name to application property in app.yaml.  
(Don't use same name, it will be fail in deploy process.)  
```
$ goapp serve  
```
Open "localhost:8080" on your web browser.  
If you watch admin server page, open "localhost:8000".
### Deploy

```
$ goapp deploy  
```

## Assets(Character Face's Set)

images/face/default/\*.png(00.png ~ 07.png)  
Painted by fuaim  
(c) Copyright 2015 fuaim  
These works are licensed under a Creative Commons Attribution-ShareAlike 2.1(or later version) JP license.  
これらの作品は Creative Commons 表示 - 継承 2.1(またはそれ以降) 日本 ライセンスの下で提供されます。  
https://creativecommons.org/licenses/by-sa/2.1/jp/

## Other Assets
A few pictures are Public Domain, Please see "images/assets_license.txt"  

If you want to use original Character Face on this app,  
1. Copy to images/face/  
2. Edit the "characters.yaml"  

# Author
Haruki Tsurumoto (tSU-RooT) <tsu.root@gmail.com>
# License
Lycos is free software.  
Programs are licensed under MIT License.  
