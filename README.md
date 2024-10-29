# http-here

Share folder via http with upload

Multiple files upload to current showed folder

Also you can download any files inside current folder, just click on them

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
wget https://github.com/western/http-here/releases/download/v1.0.4/http-here.gz
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

```console
http-here --user loginXX --password MugMf7AHs .
```


## Desktop window
<p align="center">
    <img src="https://github.com/western/http-here/blob/dev/doc/screen.png?raw=true&1" />
</p>


## Notes

> [!CAUTION]
> Be careful, if you start this App on public network interface, anybody can work with it

> [!CAUTION]  
> Always run this app only under unprivileged common user

- If you run application under some User, this user should be have privileges to write current folder


## History

### backlog
- [ ] make img thumbnail storage 

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


