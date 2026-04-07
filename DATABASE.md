# http-here database

## Goals
- `User table` for save specific person login/password
- `Log table` for save information

Database is disabled by default. So, you should run app with `--usedb` arg.

## Help information

```code

http-here --help

http-here user --help
http-here usermod --help
http-here userdel --help

http-here log --help

```

## Users (some examples)

### Generate list of persistent users

```code
http-here user --generate
```

### Print all users

```code
http-here user --list
```

### Modify (disable) one user

```code
http-here usermod --login XXXXXXX --disable
```

### Delete one user

```code
http-here userdel --login XXXXXXX
```



## Log (some examples)

### Print all log data to stdout

```code
http-here log --dump
```

