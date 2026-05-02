# http-here database

## Goals
- [`User table`](#users-some-examples) for persistent users login/password
- [`Log table`](#log-some-examples) store event info in one place
- `File table`

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

Add new specific user

```code

# add user with login (password will be generate as random)
http-here user --login vasya3000

# add user with login and password
http-here user --login vasya3001 --password pass3001

```

Or automatic generate and save list of random accounts

```code
http-here user --generate
```

Print all users

```code
http-here user --list
```

Modify

```code

# disable one user
http-here usermod --login XXXXXXX --disable

# change password
http-here usermod --login XXXXXXX --password TTTTTTTTT

```

Delete one user

```code
http-here userdel --login XXXXXXX
```



## Log (some examples)

Print all log data to stdout

```code
http-here log --dump
```


## File table

For old versions of `http-here` it was a place to store meta about files and use it for `search` or `image thumbnails`.

