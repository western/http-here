if (typeof window.$ != 'function') {
    window.$ = (query, cntx = document) => {
        const el = cntx.querySelector(query);
        if (!el) {
            console.warn(`Element ${query} not found`);
        }
        return el;
    };
}

if (typeof window.$$ != 'function') {
    window.$$ = (query, cntx = document) => {
        return Array.from(cntx.querySelectorAll(query));
    };
}

class DOM {
    static show(el) {
        if (el) el.style.display = 'block';
    }
    static hide(el) {
        if (el) el.style.display = 'none';
    }
    static addClass(el, className) {
        if (el) el.classList.add(className);
    }
    static removeClass(el, className) {
        if (el) el.classList.remove(className);
    }
    static toggleClass(el, className) {
        if (el) el.classList.toggle(className);
    }
    static setValue(el, val) {
        if (el) el.value = val;
    }
    static getValue(el) {
        return el ? el.value : '';
    }
    static setAttribute(el, attr, val) {
        if (el) el.setAttribute(attr, val);
    }
    static getAttribute(el, attr) {
        return el ? el.getAttribute(attr) : '';
    }
    static focus(el) {
        if (el) el.focus();
    }
    static eventAdd(el, type, listener) {
        if (el) el.addEventListener(type, listener);
    }
}

class Clipboard {
    constructor() {
        this.init();
        this.initGroupPaste();
        this.initGroupCopy();
        this.initGroupMove();
    }

    init() {
        let clipboard = JSON.parse(localStorage.getItem('clipboard') || '{}');

        if (clipboard && clipboard.mode && clipboard.mode == 'copy') {
            let el;
            if ((el = $('a.group_copy'))) {
                el.classList.add('active');
            }
            if ((el = $('a.group_move'))) {
                el.classList.remove('active');
            }
        }

        if (clipboard && clipboard.mode && clipboard.mode == 'move') {
            let el;
            if ((el = $('a.group_move'))) {
                el.classList.add('active');
            }
            if ((el = $('a.group_copy'))) {
                el.classList.remove('active');
            }
        }
    }

    static chk_fold_file_list() {
        let checked_lst = [];

        $$('input[type=checkbox][name=fold], input[type=checkbox][name=file]').forEach((aa) => {
            if (aa.checked) {
                checked_lst.push(aa.value);
            }
        });

        return checked_lst;
    }

    static chk_off() {
        $$('input[type=checkbox][name=fold], input[type=checkbox][name=file]').forEach((aa) => {
            aa.checked = false;
        });
    }

    initGroupPaste() {
        const a_group_paste = $('a.group_paste');
        if (a_group_paste) {
            a_group_paste.addEventListener('click', (ev) => {
                let clipboard = JSON.parse(localStorage.getItem('clipboard') || '{}');

                if (clipboard && clipboard.mode) {
                    let formData = new FormData();
                    formData.append('from_path', clipboard.from_path);
                    formData.append('to_path', location.pathname);

                    for (let a = 0; a < clipboard.list.length; a++) {
                        formData.append('name', clipboard.list[a]);
                    }

                    fetch('/api/' + clipboard.mode, {
                        method: 'POST',
                        body: formData,
                    })
                        .then((r) => r.json())
                        .then((js) => {
                            if (js && js.code == '200') {
                                localStorage.setItem('clipboard', JSON.stringify({}));

                                location.href = location.href;
                            } else if (js && js.msg) {
                                // reset clipboard on error
                                localStorage.setItem('clipboard', JSON.stringify({}));

                                let el;
                                if ((el = $('a.group_copy'))) {
                                    el.classList.remove('active');
                                }
                                if ((el = $('a.group_move'))) {
                                    el.classList.remove('active');
                                }

                                alert(js.msg);
                            }
                        })
                        .catch((err) => {
                            console.error('/api/' + clipboard.mode + ' catch err=', err);
                        });
                } else {
                    alert('Clipboard is empty');
                }
            });
        }
    }

    initGroupCopy() {
        const a_group_copy = $('a.group_copy');
        if (a_group_copy) {
            a_group_copy.addEventListener('click', (ev) => {
                let targ = ev.target;
                if (targ.tagName == 'I') {
                    targ = ev.target.parentNode;
                }

                let arr = Clipboard.chk_fold_file_list();

                if (arr.length == 0) {
                    alert('You need select something');
                    return;
                }

                localStorage.setItem('clipboard', JSON.stringify({ mode: 'copy', from_path: location.pathname, list: arr }));

                $('a.group_move').classList.remove('active');
                targ.classList.add('active');

                Clipboard.chk_off();

                const chk_table_rows = $('input[type=checkbox].head-chk');
                if (chk_table_rows) {
                    chk_table_rows.checked = false;
                }
            });
        }
    }

    initGroupMove() {
        const a_group_move = $('a.group_move');
        if (a_group_move) {
            a_group_move.addEventListener('click', (ev) => {
                let targ = ev.target;
                if (targ.tagName == 'I') {
                    targ = ev.target.parentNode;
                }

                let arr = Clipboard.chk_fold_file_list();

                if (arr.length == 0) {
                    alert('You need select something');
                    return;
                }

                localStorage.setItem('clipboard', JSON.stringify({ mode: 'move', from_path: location.pathname, list: arr }));

                $('a.group_copy').classList.remove('active');
                targ.classList.add('active');

                Clipboard.chk_off();

                const chk_table_rows = $('input[type=checkbox].head-chk');
                if (chk_table_rows) {
                    chk_table_rows.checked = false;
                }
            });
        }
    }
}

class API {
    constructor() {
        this.initHeaderChk();

        this.initDelOneElement();
        this.initDelGroup();
        this.initZipGroup();

        this.initMakeNewFile();
        this.initMakeNewFolder();

        this.initSearch();
        this.initRename();
        this.initEditors();

        this.initLegacyUploader();
    }

    static chk_fold_file_list() {
        let checked_lst = [];

        $$('input[type=checkbox][name=fold], input[type=checkbox][name=file]').forEach((aa) => {
            if (aa.checked) {
                checked_lst.push(aa.value);
            }
        });

        return checked_lst;
    }

    initHeaderChk() {
        const chk_table_rows = $('input[type=checkbox].head-chk');
        if (chk_table_rows) {
            chk_table_rows.addEventListener('click', (ev) => {
                const chk = ev.target.checked;

                $$('input[type=checkbox][name=fold]').forEach((aa) => {
                    aa.checked = chk;
                });

                $$('input[type=checkbox][name=file]').forEach((aa) => {
                    aa.checked = chk;
                });
            });
        }
    }

    initDelOneElement() {
        $$('a.del, button.del').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let el = ev.target;
                if (el.tagName == 'I') {
                    el = el.parentNode;
                }

                let name = el.dataset.name;

                if (confirm('Delete "' + name + '"?')) {
                    let formData = new FormData();
                    formData.append('name', name);

                    fetch('/api/delete', {
                        method: 'POST',
                        body: formData,
                    })
                        .then((r) => r.json())
                        .then((js) => {
                            if (js && js.code == '200') {
                                location.href = location.href;
                            } else if (js && js.msg) {
                                alert(js.msg);
                            }
                        })
                        .catch((err) => {
                            console.error('/api/delete/one catch err=', err);
                        });
                }
            });
        });
    }

    initDelGroup() {
        const a_group_del = $('a.group_del');
        if (a_group_del) {
            a_group_del.addEventListener('click', (ev) => {
                let arr = API.chk_fold_file_list();

                if (arr.length == 0) {
                    alert('You need select something');
                    return;
                }

                if (confirm('You really want to del this group?')) {
                    let formData = new FormData();

                    for (let a = 0; a < arr.length; a++) {
                        formData.append('name', arr[a]);
                    }

                    fetch('/api/delete', {
                        method: 'POST',
                        body: formData,
                    })
                        .then((r) => r.json())
                        .then((js) => {
                            if (js && js.code == '200') {
                                location.href = location.href;
                            } else if (js && js.msg) {
                                alert(js.msg);
                            }
                        })
                        .catch((err) => {
                            console.error('/api/delete catch err=', err);
                        });
                }
            });
        }
    }

    initZipGroup() {
        const a_group_zip = $('a.group_zip');
        if (a_group_zip) {
            a_group_zip.addEventListener('click', (ev) => {
                let arr = API.chk_fold_file_list();

                if (arr.length == 0) {
                    alert('You need select something');
                    return;
                }

                let formData = new FormData();
                for (let a = 0; a < arr.length; a++) {
                    formData.append('name', arr[a]);
                }

                fetch('/api/zip', {
                    method: 'POST',
                    body: formData,
                })
                    .then((r) => r.json())
                    .then((js) => {
                        if (js && js.code == '200') {
                            location.href = '/__temp/' + js.file;
                        } else if (js && js.msg) {
                            alert(js.msg);
                        }
                    })
                    .catch((err) => {
                        console.error('/api/zip catch err=', err);
                    });
            });
        }
    }

    initMakeNewFile() {
        const make_new_file = async (ev) => {
            let val = $('input[type=text]', ev.target.parentNode).value;

            if (val.length == 0) {
                alert('Please fill file name');
                return;
            }

            let formData = new FormData();
            formData.append('name', val);

            fetch('/api/file/touch', {
                method: 'POST',
                body: formData,
            })
                .then((r) => r.json())
                .then((js) => {
                    if (js && js.code == '200') {
                        let openOnlineEditor = $('#openOnlineEditor');
                        if (openOnlineEditor && !openOnlineEditor.checked) {
                            location.href = location.href;
                            return;
                        }

                        let regx = /\.(.+?)$/.exec(val);
                        if (!regx) {
                            location.href = location.href;
                            return;
                        }

                        let ext = regx.slice(1)[0];

                        let is_edit_doc = ext.match(/^(rtf|doc|docx|odt)$/i);
                        let is_edit_code = ext.match(/^(html|txt|js|css|sh|json)$/i);
                        let is_edit_md = ext.match(/^(md)$/i);

                        if (is_edit_doc) {
                            location.href = '/__doc' + location.pathname + '/' + val;
                            return;
                        }
                        if (is_edit_code) {
                            location.href = '/__code' + location.pathname + '/' + val;
                            return;
                        }
                        if (is_edit_md) {
                            location.href = '/__md' + location.pathname + '/' + val;
                            return;
                        }

                        location.href = location.href;
                        return;
                    } else if (js && js.msg) {
                        alert(js.msg);
                    }
                })
                .catch((err) => {
                    console.error('/api/file/touch catch err=', err);
                });
        };

        const mk_file_button = $('#make_file_button');
        if (mk_file_button) {
            mk_file_button.addEventListener('click', (ev) => {
                make_new_file(ev);
            });
        }

        const mk_file_input = $('#make_file_input');
        if (mk_file_input) {
            mk_file_input.addEventListener('keypress', (ev) => {
                if (ev.which == 13) {
                    make_new_file(ev);
                }
            });
        }

        $$('#make_file_dlg, #make_file_dlg2').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let file_modal = $('#make_file_modal');

                file_modal.addEventListener('shown.bs.modal', () => {
                    const inp = $('#make_file_input');
                    DOM.focus(inp);
                });

                const mkfileModal = new bootstrap.Modal(file_modal, {});
                mkfileModal.show();
            });
        });
    }

    initMakeNewFolder() {
        const make_new_folder = async (ev) => {
            let val = $('input[type=text]', ev.target.parentNode).value;

            if (val.length == 0) {
                alert('Please fill folder name');
                return;
            }

            let formData = new FormData();
            formData.append('name', val);

            fetch('/api/folder', {
                method: 'POST',
                body: formData,
            })
                .then((r) => r.json())
                .then((js) => {
                    if (js.code == '200') {
                        location.href = location.href;
                    } else if (js && js.msg) {
                        alert(js.msg);
                    }
                })
                .catch((err) => {
                    console.error('/api/folder catch err=', err);
                });
        };

        const mk_folder_button = $('#make_folder_button');
        if (mk_folder_button) {
            mk_folder_button.addEventListener('click', (ev) => {
                make_new_folder(ev);
            });
        }

        const mk_folder_input = $('#make_folder_input');
        if (mk_folder_input) {
            mk_folder_input.addEventListener('keypress', (ev) => {
                if (ev.which == 13) {
                    make_new_folder(ev);
                }
            });
        }

        $$('#make_folder_dlg, #make_folder_dlg2').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let folder_modal = $('#make_folder_modal');

                folder_modal.addEventListener('shown.bs.modal', () => {
                    const inp = $('#make_folder_input');
                    DOM.focus(inp);
                });

                const mkfolderModal = new bootstrap.Modal(folder_modal, {});
                mkfolderModal.show();
            });
        });
    }

    initSearch() {
        let search_submit = function (ev) {
            let val = $('input[type=text]', ev.target.parentNode).value;

            if (val.length == 0) {
                alert('Please fill search field');
                return;
            }

            location.href = '/__search/?s=' + val;
        };

        const search_button = $('#search_button');
        if (search_button) {
            search_button.addEventListener('click', (ev) => {
                search_submit(ev);
            });
        }

        const search_input = $('#search_input');
        if (search_input) {
            search_input.addEventListener('keypress', (ev) => {
                if (ev.which == 13) {
                    search_submit(ev);
                }
            });
        }

        $$('#search_dlg, #search_dlg2').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let modal = $('#search_modal');

                modal.addEventListener('shown.bs.modal', () => {
                    const inp = $('#search_input');
                    DOM.focus(inp);
                });

                const srchModal = new bootstrap.Modal(modal, {});
                srchModal.show();
            });
        });
    }

    initRename() {
        $$('a.rename').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let el = ev.target;

                if (el.tagName == 'I') {
                    el = el.parentNode;
                }
                //console.log('el=', el);

                $('#set_rename_orig').value = el.dataset.name;
                $('#set_rename_input').value = el.dataset.name;

                let mod = $('#rename_modal');

                mod.addEventListener('shown.bs.modal', () => {
                    const inp = $('#rename_input');
                    DOM.focus(inp);
                });

                const renameModal = new bootstrap.Modal(mod, {});
                renameModal.show();
            });
        });

        let rename_form_submit = () => {
            let orig = $('#set_rename_orig').value;
            let name = $('#set_rename_input').value;

            if (name.length == 0) {
                alert('Please fill name');
                return;
            }

            let formData = new FormData();
            formData.append('name', orig);
            formData.append('to', name);

            fetch('/api/rename', {
                method: 'POST',
                body: formData,
            })
                .then((r) => r.json())
                .then((js) => {
                    if (js.code == '200') {
                        location.href = location.href;
                    } else if (js && js.msg) {
                        alert(js.msg);
                    }
                })
                .catch((err) => {
                    console.error('/api/rename catch err=', err);
                });
        };

        const set_rename_button = $('#set_rename_button');
        if (set_rename_button) {
            set_rename_button.addEventListener('click', (ev) => {
                rename_form_submit(ev);
            });
        }

        const set_rename_input = $('#set_rename_input');
        if (set_rename_input) {
            set_rename_input.addEventListener('keypress', (ev) => {
                if (ev.which == 13) {
                    rename_form_submit(ev);
                }
            });
        }
    }

    initEditors() {
        $$('a.edit_code').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let el = ev.target;

                if (el.tagName == 'I') {
                    el = el.parentNode;
                }

                location.href = '/__code' + location.pathname + '/' + el.dataset.name;
            });
        });

        $$('a.edit_doc').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let el = ev.target;

                if (el.tagName == 'I') {
                    el = el.parentNode;
                }

                location.href = '/__doc' + location.pathname + '/' + el.dataset.name;
            });
        });

        $$('a.edit_md').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let el = ev.target;

                if (el.tagName == 'I') {
                    el = el.parentNode;
                }

                location.href = '/__md' + location.pathname + '/' + el.dataset.name;
            });
        });

        $$('a.player').forEach((aa) => {
            aa.addEventListener('click', (ev) => {
                let el = ev.target;

                if (el.tagName == 'I') {
                    el = el.parentNode;
                }

                location.href = '/__player' + location.pathname + '/' + el.dataset.name;
            });
        });
    }

    initLegacyUploader() {
        $$('#upload_file, #upload_file2').forEach((upl) => {
            upl.addEventListener('change', this.evTargetFiles.bind(this));
        });
    }

    evTargetFiles(ev) {
        const files = ev.target.files;

        if (files.length > config.files_count_max) {
            alert(`Count of files is more than ${config.files_count_max}.`);
            location.href = location.href;
            return;
        }

        let formData = new FormData();

        Array.prototype.forEach.call(files, function (file) {
            if (file.size > config.fieldSize_max) {
                alert('File "' + file.name + `" size is overload "${config.fieldSize_max_human}"`);
            } else {
                formData.append('fileBlob', file, encodeURIComponent(file.name));
            }
        });

        let submit = async function () {
            $('#progress').setAttribute('max', 100);
            $('#progress').setAttribute('value', 0);
            $('#progress').style.display = 'block';

            let xhr = new XMLHttpRequest();

            xhr.upload.addEventListener(
                'progress',
                function (ev, th) {
                    if (ev.lengthComputable) {
                        $('#progress').setAttribute('max', ev.total);
                        $('#progress').setAttribute('value', ev.loaded);
                    }
                },
                false,
            );

            xhr.onreadystatechange = function (ev) {
                if (xhr.readyState == 4 && ev.target.status == 500) {
                    let json = JSON.parse(ev.target.responseText);
                    alert(json.msg);
                }

                if (xhr.readyState == 4 && ev.target.status == 200) {
                    location.href = location.href;
                }
            };
            xhr.open('POST', '/api/file');
            xhr.send(formData);
        };

        submit();
    }
}

// --------------------------------------------------------------------------------------------------------------------------------------

class BaseClient {
    constructor(args = {}) {
        this.prefixUrl = args.prefixUrl || '/api/file';

        this.retryDelay = args.retryDelay || 1000;
        this.maxRetries = args.maxRetries || 2;
        this.requestTimeout = args.requestTimeout || 5000;
        if (args.requestTimeout === 0) {
            this.requestTimeout = 0;
        }
        this.sessions = new Map();

        this.sizeSum = 0;
        this.sizeProgress = 0;

        this.onProgress = args.onProgress || (() => {});
        this.onComplete = args.onComplete || (() => {});
        this.onError = args.onError || (() => {});
        this.onFailed = args.onFailed || (() => {});
    }

    async fetchWithTimeout(url, options, timeout) {
        // without timeout
        if (timeout == 0) {
            return fetch(url, options);
        }

        // with
        return Promise.race([
            fetch(url, options),
            new Promise((_, reject) => setTimeout(() => reject(new Error('fetchWithTimeout timeout ' + timeout)), timeout)),
        ]);
    }

    humanFileSize(bytes) {
        if (bytes === 0) return '0 Bytes';
        const k = 1024;
        const sizes = ['Bytes', 'KiB', 'MiB', 'GiB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    waitTime(ms) {
        return new Promise((resolve) => setTimeout(resolve, ms));
    }

    generateFileId(file) {
        return `${file.name}--${file.size}--${file.lastModified}`;
    }

    // --------------------------------------------------------------------------------------------------------------------------------------
    // ENGINE

    async prepareAndUploadFile(file) {}

    async uploadFile(file) {}

    // --------------------------------------------------------------------------------------------------------------------------------------
    // INPUT LISTENER

    on(nme, cb) {
        switch (nme.toLowerCase()) {
            case 'progress':
                this.onProgress = cb;
                break;
            case 'complete':
                this.onComplete = cb;
                break;
            case 'error':
                this.onError = cb;
                break;
            case 'failed':
                this.onFailed = cb;
                break;
        }
    }

    async handleInputListener(ev) {
        let files = ev.target.files;
        let promises = [];

        this.sizeSum = 0;
        this.sizeProgress = 0;
        this.onProgress('open');

        for (let a = 0; a < files.length; a++) {
            promises.push(this.prepareAndUploadFile(files[a]));
        }

        await Promise.all(promises);

        this.onProgress('close');
    }

    setupEventListener(qu) {
        const fileInputs = document.querySelectorAll(qu);
        fileInputs.forEach((input) => {
            input.addEventListener('change', this.handleInputListener.bind(this));
        });
    }
}

// --------------------------------------------------------------------------------------------------------------------------------------

class SimpleUploadClient extends BaseClient {
    constructor(args = {}) {
        super(args);

        this.prefixUrl = args.prefixUrl || '/api/file';
    }

    // --------------------------------------------------------------------------------------------------------------------------------------
    // ENGINE

    async prepareAndUploadFile(file) {
        const fileId = this.generateFileId(file);
        const sessionInfo = {
            fileId,
            name: file.name,
            size: file.size,
            sizeHuman: this.humanFileSize(file.size),
            errors: [],
            isDelayed: false,
            isFailed: false,
            isDone: false,
            startTime: Date.now(),
            endTime: undefined,
        };

        this.sizeSum += file.size;

        this.sessions.set(fileId, sessionInfo);

        return this.uploadFile(file);
    }

    async uploadFile(file) {
        const fileId = this.generateFileId(file);

        let formData = new FormData();
        formData.append('fileBlob', file);
        formData.append('path', location.pathname);

        for (let attempt = 0; attempt < this.maxRetries; attempt++) {
            let sessionInfo = this.sessions.get(fileId);

            if (sessionInfo.isDone) {
                continue;
            }

            if (sessionInfo.isDelayed) {
                await this.waitTime(this.retryDelay);
            }

            await this.fetchWithTimeout(
                this.prefixUrl,
                {
                    method: 'POST',
                    body: formData,
                },
                this.requestTimeout,
            )
                .then((r) => r.text())
                .then((text) => {
                    try {
                        const js = JSON.parse(text);

                        if (js && js.code == 200) {
                            sessionInfo.isDone = true;
                            sessionInfo.endTime = Date.now();

                            this.sizeProgress += sessionInfo.size;
                            this.onProgress('process_done', sessionInfo, this.sizeSum, this.sizeProgress);

                            this.sessions.set(fileId, sessionInfo);
                            this.onComplete(sessionInfo, js);
                        } else if (js && js.msg) {
                            throw new Error(js.code + ' ' + js.msg);
                        }
                    } catch (err) {
                        //console.error('fetch TRY err=', err);
                        this.onError(err);

                        sessionInfo.isDelayed = true;
                        sessionInfo.errors.push(err);
                        this.sessions.set(fileId, sessionInfo);
                    }
                })
                .catch((err) => {
                    //console.error('fetch CATCH err=', err);
                    this.onError(err);

                    sessionInfo.isDelayed = true;
                    sessionInfo.errors.push(err);
                    this.sessions.set(fileId, sessionInfo);
                });
        }

        let sessionInfo = this.sessions.get(fileId);
        if (!sessionInfo.endTime) {
            sessionInfo.endTime = Date.now();
        }
        if (!sessionInfo.isDone) {
            sessionInfo.isFailed = true;
        }
        if (!sessionInfo.isDone) {
            this.onFailed(sessionInfo);
        }

        this.sessions.set(fileId, sessionInfo);
    }
}

(() => {
    let uploadClient = new SimpleUploadClient({
        prefixUrl: '/api/file',
        retryDelay: 2000,
        maxRetries: 2,
        requestTimeout: 0, // disable timeout
        onProgress: (mode, sessionInfo, sizeSum, sizeProgress) => {
            //console.info('onProgress', mode);

            if (mode == 'open') {
                $('#progress').setAttribute('max', 100);
                $('#progress').setAttribute('value', 0);
                $('#progress').style.display = 'block';
            }

            if (mode == 'process_done') {
                $('#progress').setAttribute('max', sizeSum);
                $('#progress').setAttribute('value', sizeProgress);
            }

            if (mode == 'close') {
                $('#progress').style.display = 'none';
            }
        },
        onComplete: (sessionInfo, js) => {
            console.info('onComplete', sessionInfo.name, js);

            if (js && js.code == 200) {
                //location.href = location.href;
            }
        },
        onError: (err) => {
            console.info('onError', err.name, err.message);
        },
        onFailed: (sessionInfo) => {
            console.info('onFailed', sessionInfo);
        },
    });

    //uploadClient.setupEventListener('#upload_file, #upload_file2');
})();

// --------------------------------------------------------------------------------------------------------------------------------------

class PartUploadClient extends BaseClient {
    constructor(args = {}) {
        super(args);

        this.prefixUrl = args.prefixUrl || '/api/file/part';

        this.chunkSize = args.chunkSize || 8 * 1024 * 1024;
    }

    getFileChunks(file) {
        const fileId = this.generateFileId(file);
        const chunks = [];
        let start = 0;
        let counter = 0;

        while (start < file.size) {
            const end = Math.min(start + this.chunkSize, file.size);
            const data = file.slice(start, end);
            chunks.push({
                n: counter,
                data,
                fileId,
                name: file.name,
                size: data.size,
                sizeHuman: this.humanFileSize(data.size),
                errors: [],
                isDelayed: false,
                isFailed: false,
                isDone: false,
                startTime: undefined,
                endTime: undefined,
            });
            start = end;
            counter++;
        }

        return chunks;
    }

    // --------------------------------------------------------------------------------------------------------------------------------------
    // ENGINE

    async prepareAndUploadFile(file) {
        const fileId = this.generateFileId(file);

        const chunks = this.getFileChunks(file);

        const sessionInfo = {
            fileId,
            name: file.name,
            size: file.size,
            size_human: this.humanFileSize(file.size),
            chunks,
            uploadedChunks: 0,
            errors: [],
            to_path: location.pathname,
            startTime: Date.now(),
            endTime: undefined,
        };

        this.sessions.set(fileId, sessionInfo);

        this.sizeSum += file.size;

        await this.uploadFile(fileId, sessionInfo);

        console.log('prepareAndUploadFile end');
    }

    async uploadFile(fileId, sessionInfo) {
        let promises = [];

        for (let a = 0; a < sessionInfo.chunks.length; a++) {
            let pr = this.uploadChunk(fileId, a);
            promises.push(pr);
        }

        await Promise.all(promises);

        return this.uploadFileDone(fileId);
    }

    async uploadChunk(fileId, chunk_index) {
        let sessionInfo = this.sessions.get(fileId);
        let chunkInfo = sessionInfo.chunks[chunk_index];

        chunkInfo.startTime = Date.now();
        sessionInfo.chunks[chunk_index] = chunkInfo;

        let formData = new FormData();
        formData.append('file_id', sessionInfo.fileId);
        formData.append('chunk_index', chunk_index);
        formData.append('data', chunkInfo.data);

        for (let attempt = 0; attempt < this.maxRetries; attempt++) {
            let chunkInfo = sessionInfo.chunks[chunk_index];

            if (chunkInfo.isDone) {
                continue;
            }

            if (chunkInfo.isDelayed) {
                await this.waitTime(this.retryDelay);
            }

            await this.fetchWithTimeout(
                this.prefixUrl,
                {
                    method: 'POST',
                    body: formData,
                },
                this.requestTimeout,
            )
                .then((r) => r.text())
                .then((text) => {
                    try {
                        const js = JSON.parse(text);

                        if (js && js.code == 200) {
                            chunkInfo.isDone = true;
                            chunkInfo.endTime = Date.now();
                            sessionInfo.chunks[chunk_index] = chunkInfo;

                            this.sizeProgress += chunkInfo.size;
                            this.onProgress('process_done', chunkInfo, this.sizeSum, this.sizeProgress);

                            //this.onComplete(chunkInfo, js);
                        } else if (js && js.msg) {
                            throw new Error(js.code + ' ' + js.msg);
                        }
                    } catch (err) {
                        //console.error('fetch TRY err=', err);
                        this.onError(err);

                        chunkInfo.isDelayed = true;
                        chunkInfo.errors.push(err);
                        sessionInfo.chunks[chunk_index] = chunkInfo;
                    }
                })
                .catch((err) => {
                    //console.error('fetch CATCH err=', err);
                    this.onError(err);

                    chunkInfo.isDelayed = true;
                    chunkInfo.errors.push(err);
                    sessionInfo.chunks[chunk_index] = chunkInfo;
                });
        }

        chunkInfo = sessionInfo.chunks[chunk_index];
        if (!chunkInfo.endTime) {
            chunkInfo.endTime = Date.now();
        }
        if (!chunkInfo.isDone) {
            chunkInfo.isFailed = true;
        }
        if (!chunkInfo.isDone) {
            this.onFailed(sessionInfo, chunkInfo);
        }

        sessionInfo.chunks[chunk_index] = chunkInfo;
    }

    async uploadFileDone(fileId) {
        let sessionInfo = this.sessions.get(fileId);

        let formData = new FormData();
        formData.append('file_id', fileId);
        formData.append('name', sessionInfo.name);
        formData.append('to_path', sessionInfo.to_path);
        formData.append('chunk_cnt', sessionInfo.chunks.length);

        return fetch(this.prefixUrl + '/done', {
            method: 'POST',
            body: formData,
        })
            .then((r) => r.text())
            .then((text) => {
                try {
                    const js = JSON.parse(text);

                    if (js && js.code == 200) {
                        sessionInfo.isDone = true;
                        sessionInfo.endTime = Date.now();

                        //this.sizeProgress += sessionInfo.size;
                        //this.onProgress('process_done', sessionInfo, this.sizeSum, this.sizeProgress);

                        this.sessions.set(fileId, sessionInfo);
                        this.onComplete(sessionInfo, js);
                    } else if (js && js.msg) {
                        throw new Error(js.code + ' ' + js.msg);
                    }
                } catch (err) {
                    //console.error('fetch TRY err=', err);
                    this.onError(err);

                    //sessionInfo.isDelayed = true;
                    sessionInfo.errors.push(err);
                    this.sessions.set(fileId, sessionInfo);
                } finally {
                    let sessionInfo = this.sessions.get(fileId);
                    if (!sessionInfo.endTime) {
                        sessionInfo.endTime = Date.now();
                    }
                    if (!sessionInfo.isDone) {
                        sessionInfo.isFailed = true;
                    }
                    if (!sessionInfo.isDone) {
                        this.onFailed(sessionInfo);
                    }

                    this.sessions.set(fileId, sessionInfo);
                }
            })
            .catch((err) => {
                //console.error('fetch CATCH err=', err);
                this.onError(err);

                //sessionInfo.isDelayed = true;
                sessionInfo.errors.push(err);
                this.sessions.set(fileId, sessionInfo);
            });
    }
}

(() => {
    let uploadClient = new PartUploadClient({
        chunkSize: 25 * 1024 * 1024,
        retryDelay: 2000,
        maxRetries: 2,
        requestTimeout: 0, // disable timeout
        onProgress: (mode, chunkInfo, sizeSum, sizeValue) => {
            //console.info('onProgress', mode);

            if (mode == 'open') {
                $('#progress').setAttribute('max', 100);
                $('#progress').setAttribute('value', 'xxx');
                $('#progress').style.display = 'block';
            }

            if (mode == 'process_done') {
                $('#progress').setAttribute('max', sizeSum);
                $('#progress').setAttribute('value', sizeValue);
            }

            if (mode == 'close') {
                $('#progress').style.display = 'none';
            }
        },
        onComplete: (sessionInfo, js) => {
            console.info('onComplete', sessionInfo.name, js);

            if (js && js.code == 200) {
                //location.href = location.href;
            }
        },
        onError: (err) => {
            console.info('onError', err.name, err.message);
        },
        onFailed: (sessionInfo, chunkInfo) => {
            console.info('onFailed', sessionInfo, chunkInfo);
        },
    });

    //uploadClient.setupEventListener('#upload_file, #upload_file2');
})();

// --------------------------------------------------------------------------------------------------------------------------------------

(() => {
    const clipboard = new Clipboard();
    const api = new API();
})();
