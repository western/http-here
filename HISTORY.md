
## History



### 1.11.4
- [x] redirect if you try open folder during SPA application mode
- [x] fix refresh doc thumbnails after edit
- [x] change layouts for errors
- [x] project code restructuring

### 1.11.3
- [x] run WalkAndTreeBuild only for extend mode
- [x] change view for default mode
- [x] fix for mobile upload

### 1.11.0
- [x] database support
- [x] "make new file" button and API
- [x] search button and API
- [x] first TLS key files generate without verbosity
- [x] body limit up to 14 GB
- [x] max upload files up to 100
- [x] add SPA version
- [x] new API:  `/api/list`  `/__convert/filename.docx`  `/api/search`
- [x] use clipboard for SPA client group operations


<hr>

### 1.9.2
- [x] show file_name while file edit
- [x] change online edit templates
- [x] change codemirror

### 1.9.0
- [x] add source code editor for `html|txt|js|css|md`

### 1.8.1
- [x] enable preview for `rtf|doc|docx|odt`

### 1.8.0
- [x] online editor for `html|rtf|doc|docx|odt`

### 1.7.0
- [x] top buttons operations
- [x] api/copy
- [x] api/rename
- [ ] "edit" still planning
- [ ] "share" still planning


<hr>

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

