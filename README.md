# http-here

Share folder via http with upload

Multiple files upload to current showed folder

Also you can download any files inside current folder, just click on them

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/screen.png?raw=true&1" />
</p>

## Mobile screen

<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/screen_mobile.png?raw=true&1" />
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
wget https://github.com/western/http-here/releases/download/v1.0.9/http-here.gz
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

## The safest

```console
http-here --tls --basic /path/to/you
```
read for TLS support below

## Only share

```console
http-here --upload-disable --folder-make-disable /tmp/fold
```



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

## History

### backlog
- [ ] make img thumbnail storage


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
- upload file up to 7 Gb
- multiple upload to 20 files
- make folder in current show path
- show current folder
- basic auth for one account


