# http-here

Simple zero-configuration command line http server with lightweight interface to work with files


[![Go Report Card](https://goreportcard.com/badge/github.com/western/http-here?style=flat-square)](https://goreportcard.com/report/github.com/western/http-here)
[![Go Reference](https://pkg.go.dev/badge/github.com/western/http-here.svg)](https://pkg.go.dev/github.com/western/http-here)
[![Releases](https://img.shields.io/github/release/western/http-here/all.svg?style=flat-square)](https://github.com/western/http-here/releases)
[![LICENSE](https://img.shields.io/github/license/western/http-here.svg?style=flat-square)](https://github.com/western/http-here/blob/dev/LICENSE.txt)


> Share folder via http with upload

> Multiple files upload to current showed folder

> In extended mode you can doing more

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/screen-top6.png?raw=true" />
</p>

## Mobile screen

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/screen_mobile4.png?raw=true" />
</p>





## Install from github

```console
git clone --branch v1.14.3 https://github.com/western/http-here
cd http-here
make build
```
or
```console
go install github.com/western/http-here/cmd/http-here@v1.14.3
```



## Manual download

linux / amd64

```console
# go to your home bin
cd ~/bin

# download and unpack
wget https://github.com/western/http-here/releases/download/v1.14.3/http-here.gz
gzip -d http-here.gz

chmod +x http-here
```

windows / amd64
```console
https://github.com/western/http-here/releases/download/v1.14.3/http-here.zip
```


## Run
```console
http-here /tmp
```
or
```console
http-here --port 7999 /path/to/folder
```


## If you switch --extend-mode

```console
http-here --extend-mode /tmp
```

App will change main list view to table. And you can operate with files - delete, move, copy

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/panel_buttons2.png?raw=true"  >
</p>

Below you see display width more than 992 pix (1), less than (2) and mobile window (3):

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/width_screen_compare10.png?raw=true"  >
    <img src="https://github.com/western/http-here/blob/dev/doc/width_screen_compare11.png?raw=true"  >
</p>

> [!IMPORTANT]  
> During group operations COPY or MOVE all target files/folders will be rewrite

## Preview doc button

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/preview_doc_button.png?raw=true"  >
    
</p>

For [Thumbnails support](#thumbnails-support) read below




## Basic auth

> [!IMPORTANT]  
> It is recommend for work on public network interfaces

every time when you start, you get a list of random accounts

```console
http-here --basic .
```

or only one basic auth specific user

```console
http-here --login loginXX --password MugMf7AHs .
```

## The safest run

```console
http-here --tls --basic /path/to/you
```

read for [TLS Support](#automatic-tls-keys-generate) below

## Only share

```console
http-here --share-only /tmp/fold
```

## Run with prefork

Prefork help to handle with multiple heavy query (big image gallery as example)

```console
http-here --prefork --extend-mode /tmp
```



## Online editor

You can online edit files `html, rtf, doc, docx, odt` as office files.

Or `html, txt, js, css, md` formats as source code.

Or `md` as markdown.

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/online_editor_cmp2.png?raw=true"  />
    
</p>

You need `libreoffice` package for office files.

Office files follow this flow: `file.doc => file.html, edit => file.doc`

## File encrypt

> [!IMPORTANT]  
> Be careful. If you download `.crypt` file with WRONG password, it file will be contain MESS of bytes

<br>

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/code_to_encrypt5.png?raw=true" width="60%" >
    
</p>

Your server need package `openssl`. It will be use `openssl aes-256-cbc`

```console
http-here --extend-mode --crypt /tmp
```

Then, set your passcode to the form. The passcode store on the form between requests and you not need input it every time (if you clear it server will not use openssl).

During the process of uploading, your files will be encrypt and their EXT change to `.crypt`

When files lying on your server, their data is encrypted.

If you need decrypt any `.crypt` flles, set your passcode, and click on file. During download this file, it will be decrypt on the fly.

### Server will be encrypt upload file:
```console
http-here --extend-mode --crypt /tmp
```
- if you set `--crypt` arg on cmd
- if you set passcode (pass code set by form)

### Server will be decrypt download file:
```console
http-here --extend-mode --crypt /tmp
```
- if you set `--crypt` arg on cmd
- if filename contain `.crypt` extension
- if you set right passcode (pass code set by form)

### Server will be decrypt download file (case 2):
```console
http-here /tmp
```
- if filename contain `.crypt` extension
- if you get file with `code` param: `/fold3/file.jpg.crypt?code=YOUR_PASS_HERE`



## Notes

> [!CAUTION]
> Be careful, if you start this App on public network interface, anybody can work with it

> [!CAUTION]  
> Always run this app only under unprivileged common user

- If you run application under some User, this user should be have privileges to write current folder

## Automatic TLS keys generate

- When you start server with `--tls` option, all keys generate automatically

```console
http-here --tls .
```

- Server use self signed certs, generated at first time. Thus you need approve this connection on your clients.

<p float="left">
  <img src="https://github.com/western/http-here/blob/dev/doc/chrome_self_signed_cert.png?raw=true" width="45%" >
  <img src="https://github.com/western/http-here/blob/dev/doc/firefox_self_signed_cert.png?raw=true" width="45%" >
</p>

## Magic file index.html inside any folder

If you put inside folder file `index.html`, it will be return as context

## Thumbnails support

For document preview you need `libreoffice` package. Formats `pdf, rtf, doc, docx, xls, xlsx, odt, ods`.

For `jpg, gif, png` you not need anything.

All previews generate during first view. One time.

## Linux packages needs for full functional


- `libreoffice` - for doc thumbnails, for doc files online edit
- `openssl` - encrypt file support

## API

[western/http-here/refs/heads/dev/internal/docs/openapi.yaml](https://editor.swagger.io/?url=https://raw.githubusercontent.com/western/http-here/refs/heads/dev/internal/docs/openapi.yaml)

## You can ask any question or suggest something

https://github.com/western/http-here/issues

## History

### backlog
- [ ] add --log and --tee args for save output (or database?)
- [ ] change background actions for FS drivers (i need one abstraction layer)
- [ ] problem: how decide to run md5sum inside some folder
- [ ] database, separate branch without?
- [ ] tests
- [ ] make builder for configure app compiler? as example - add all libraries to local assets, jquery, bootstrap, bootstrap-icons
- [ ] prepare frontend react code
- [ ] REST restructure

### 1.14.0
- [x] add --usedb param (database disabled by default)
- [x] cache readdir and --cache-dir param for timeout
- [x] server connections optimization
- [x] add --share-only param
- [x] remove jquery
- [x] disable WalkAndTreeBuild2, remove bstreeview (rewrite to "clipboard style")
- [x] add MARKDOWN editor (without SPA app)
- [x] refresh SPA app - change upload api and checkbox reset
- [x] rewrite code





### [other history here](HISTORY.md)

## Mascot

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/mascot.jpg?raw=true" width="40%"  />
</p>

## Pirates hiding their http

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/pirates_hiding_their_http.jpg?raw=true" />
</p>


