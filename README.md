# http-here

Share folder via http with upload

Multiple files upload to current showed folder

In extended mode you can delete or download group of files

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/screen.png?raw=true&1" />
</p>

## Mobile screen

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/screen_mobile.png?raw=true&1" />
</p>

## If you switch --extend-mode

App will change main list view to table.

Below you see display width more than 992 pix (1), less than (2) and mobile window (3):

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/width_screen_compare5.png?raw=true"  >
    <img src="https://github.com/western/http-here/blob/dev/doc/width_screen_compare6.png?raw=true"  >
</p>

## Install from github
```console
go install github.com/western/http-here
```

## Manual download

linux / amd64

```console
# go to your home bin
cd ~/bin

# download and unpack
wget https://github.com/western/http-here/releases/download/v1.6.1/http-here.gz
gzip -d http-here.gz

chmod +x http-here
```

## Run
```console
http-here /tmp
```
or
```console
http-here --port 7999 /path/to/folder
```

## Basic auth

> [!IMPORTANT]  
> It is recommend for work on public network interfaces

every time when you start, you get a list of random accounts

```console
http-here --basic .
```

or only one basic auth specific user

```console
http-here --user loginXX --password MugMf7AHs .
```

## The safest run

```console
http-here --tls --basic /path/to/you
```
read for TLS support below

## Only share

```console
http-here --upload-disable --folder-make-disable /tmp/fold
```

## Run with prefork

Prefork help to handle with multiple heavy query (big image gallery as example)

```console
http-here --prefork --extend-mode /tmp
```

If you run --prepare-thumbnails one time you maybe not need prefork

```console
http-here --prepare-thumbnails --extend-mode /tmp
```

## File encrypt

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/code_to_encrypt2.png?raw=true"  >
    
</p>

Your server need package `openssl`

```console
http-here --extend-mode --crypt /tmp
```

Then, set your passcode to the form.

During the process of uploading, your files will be encrypt and their EXT change to `.crypt`

When files lying on your server, their data is crypted.

If you need decrypt any `.crypt` flles, set your passcode, and click on file. During download this file, it will be decrypt on the fly.

### Server will be crypt upload file:
```console
http-here --extend-mode --crypt /tmp
```
- if you set `--crypt` arg on cmd
- if you set passcode (pass code send by form input)

### Server will be decrypt download file:
```console
http-here --extend-mode --crypt /tmp
```
- if you set `--crypt` arg on cmd
- if filename contain `.crypt` extension
- if you set right passcode (pass code send by form input)

### Server will be decrypt download file (case 2):
```console
http-here /tmp
```
- if filename contain `.crypt` extension
- if you get file with `code` param: `/fold3/file.jpg.crypt?code=YOUR_PASS_HERE`

Server will be use `openssl aes-256-cbc`

## Notes

> [!CAUTION]
> Be careful, if you start this App on public network interface, anybody can work with it

> [!CAUTION]  
> Always run this app only under unprivileged common user

- If you run application under some User, this user should be have privileges to write current folder

## Automatic TLS keys generate

- For start HTTPS server you need `easyrsa` linux package
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

If you put inside folder file index.html, it will be return as context



## You can ask any question or suggest something

https://github.com/western/http-here/issues

## History

### backlog
- [ ] add --log and --tee args for save output
- [ ] change background actions for FS drivers (i need one abstraction layer)

### 1.6.0
- [x] add file crypt support

need openssl package

### 1.5.7
- [x] add experimental preview office files

you need libreoffice package 

### 1.5.6
- [x] fix zero size thumbnail file

### 1.5.1 - 1.5.4
- [x] add --prepare-thumbnails key (Run and make thumbnails for target folders)
- [x] add compression for server output (Change page size)
- [x] template fix width

### 1.5.0
- [x] add prefork arg

Prefork help to handle with multiple heavy query (big image gallery as example)

### 1.4.0
- [x] add sort option

folders and files sort separately

### 1.3.2
- [x] enable thumb folder control (remove files older than 30 days)
- [x] change hash_name build (in case moving between folders it should be better)
- [x] show move_to button for thumbnails mode too

### 1.3.0
- [x] added group of files move function (with folder select panel)
- [x] added clear cache function, but disabled

### 1.2.0
- [x] make img thumbnail storage

### 1.1.0
- [x] add extended view mode (replace delete mode)

### 1.0.11
- [x] modify api for several names support
- [x] temporary storage support .httphere/temp
- [x] add group operations: delete and zip

### 1.0.10
- [x] add --delete-enable arg
- [x] check index file index.html inside folder and show it
- [x] handle some 500-x errors

### 1.0.9
- [x] add TLS

### 1.0.8
- [x] add --basic arg

### 1.0.7
- [x] add arg index-disable
- [x] use Locals instead setenv
- [x] change err handlers

### 1.0.6
- [x] show extended info
- [x] fix datetime in log

### 1.0.5
- [x] add --upload-disable and --folder-make-disable cmd keys
- [x] fix read errors
- [x] add cmd color

### 1.0.4
- add clear values and check exists

### 1.0.3
- [x] clear all income variables
- [x] rewrite log info

### 1.0.0
first release
- upload file BodyLimit 7 Gb
- multiple upload to 20 files
- make folder in current show path
- show current folder
- basic auth for one account

## Pirates hiding their http

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/pirates_hiding_their_http.jpg?raw=true" />
</p>


