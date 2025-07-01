/*
const query = document.querySelector.bind(document);
const queryAll = document.querySelectorAll.bind(document);

const $ = document.querySelector.bind(document);
const $$ = document.querySelectorAll.bind(document);
*/

if (typeof window.$$ != 'function') {
    window.$$ = (arg0, arg1) => {
        if (arg0 && arg1) {
            return arg0.querySelectorAll(arg1);
        }

        if (arg0 && !arg1) {
            return document.querySelectorAll(arg0);
        }
    };
}

if (typeof window.$ != 'function') {
    window.$ = (arg0, arg1) => {
        if (arg0 && arg1) {
            return arg0.querySelector(arg1);
        }

        if (arg0 && !arg1) {
            return document.querySelector(arg0);
        }
    };
}

// --------------------------------------------------------------------------------------------------------------------------------------


const getCookie = (name) => {
    let matches = document.cookie.match(new RegExp(
        "(?:^|; )" + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
    ));
    return matches ? decodeURIComponent(matches[1]) : undefined;
}

const setCookie = (name, value, options = {}) => {

    options = {
        path: '/',
        ...options
    };

    if (options.expires instanceof Date) {
        options.expires = options.expires.toUTCString();
    }

    let updatedCookie = encodeURIComponent(name) + "=" + encodeURIComponent(value);

    for (let optionKey in options) {
        updatedCookie += "; " + optionKey;
        let optionValue = options[optionKey];
        if (optionValue !== true) {
            updatedCookie += "=" + optionValue;
        }
    }

    document.cookie = updatedCookie;
}


// setCookie('user', 'John', {secure: true, 'max-age': 3600});


// --------------------------------------------------------------------------------------------------------------------------------------

(() => {
    //console.log('DOM ready?');

    // --------------------------------------------------------------------------------------------------------------------------------------

    const chk_fold_file_list = () => {
        let checked_lst = [];

        $$('input[type=checkbox][name=fold], input[type=checkbox][name=file]').forEach((aa) => {
            if (aa.checked) {
                checked_lst.push(aa.value);
            }
        });

        return checked_lst;
    };

    const chk_off = () => {
        $$('input[type=checkbox][name=fold], input[type=checkbox][name=file]').forEach((aa) => {
            aa.checked = false;
        });
    };

    // --------------------------------------------------------------------------------------------------------------------------------------

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
        //$('a.group_move').classList.add("active");
        //$('a.group_copy').classList.remove("active");

        let el;
        if ((el = $('a.group_move'))) {
            el.classList.add('active');
        }
        if ((el = $('a.group_copy'))) {
            el.classList.remove('active');
        }
    }

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
                        if (js.code == '200') {
                            localStorage.setItem('clipboard', JSON.stringify({}));

                            location.href = location.href;
                        } else {
                            if (js.msg) {
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
                        }
                    })
                    .catch((err) => {
                        console.error(err);
                    });
            } else {
                alert('Clipboard is empty');
            }
        });
    }

    const a_group_copy = $('a.group_copy');
    if (a_group_copy) {
        a_group_copy.addEventListener('click', (ev) => {
            let targ = ev.target;
            if (targ.tagName == 'I') {
                targ = ev.target.parentNode;
            }

            let arr = chk_fold_file_list();

            if (arr.length == 0) {
                alert('You need select something');
                return;
            }

            localStorage.setItem('clipboard', JSON.stringify({ mode: 'copy', from_path: location.pathname, list: arr }));

            $('a.group_move').classList.remove('active');
            targ.classList.add('active');

            chk_off();

            const chk_table_rows = $('input[type=checkbox].head-chk');
            if (chk_table_rows) {
                chk_table_rows.checked = false;
            }
        });
    }

    const a_group_move = $('a.group_move');
    if (a_group_move) {
        a_group_move.addEventListener('click', (ev) => {
            let targ = ev.target;
            if (targ.tagName == 'I') {
                targ = ev.target.parentNode;
            }

            let arr = chk_fold_file_list();

            if (arr.length == 0) {
                alert('You need select something');
                return;
            }

            localStorage.setItem('clipboard', JSON.stringify({ mode: 'move', from_path: location.pathname, list: arr }));

            $('a.group_copy').classList.remove('active');
            targ.classList.add('active');

            chk_off();

            const chk_table_rows = $('input[type=checkbox].head-chk');
            if (chk_table_rows) {
                chk_table_rows.checked = false;
            }
        });
    }

    // --------------------------------------------------------------------------------------------------------------------------------------

    $$('a.del, button.del').forEach((aa) => {
        aa.addEventListener('click', (ev) => {
            let el = ev.target;
            if (el.tagName == 'I') {
                el = el.parentNode;
            }
            //console.log('el=', el);

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
                        if (js.code == '200') {
                            location.href = location.href;
                        } else {
                            if (js.msg) {
                                alert(js.msg);
                            }
                        }
                    })
                    .catch((err) => {
                        console.error(err);
                    });
            }
        });
    });

    const a_group_del = $('a.group_del');
    if (a_group_del) {
        a_group_del.addEventListener('click', (ev) => {
            let arr = chk_fold_file_list();

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
                        if (js.code == '200') {
                            location.href = location.href;
                        } else {
                            if (js.msg) {
                                alert(js.msg);
                            }
                        }
                    })
                    .catch((err) => {
                        console.error(err);
                    });
            }
        });
    }

    // --------------------------------------------------------------------------------------------------------------------------------------

    const a_group_zip = $('a.group_zip');
    if (a_group_zip) {
        a_group_zip.addEventListener('click', (ev) => {
            let arr = chk_fold_file_list();

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
                    if (js.code == '200') {
                        
                        location.href = '/__temp/' + js.file;
                        
                    } else {
                        if (js.msg) {
                            alert(js.msg);
                        }
                    }
                })
                .catch((err) => {
                    console.error(err);
                });
        });
    }

    // --------------------------------------------------------------------------------------------------------------------------------------

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

    // --------------------------------------------------------------------------------------------------------------------------------------

    const make_new_file = async (ev) => {
        let val = $(ev.target.parentNode, 'input[type=text]').value;

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
                if (js.code == '200') {
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
                    let is_edit_code = ext.match(/^(html|txt|js|css)$/i);
                    let is_edit_md = ext.match(/^(md)$/i);

                    if (is_edit_doc) {
                        location.href = '/__doc' + location.pathname + '/' + val;
                    }
                    if (is_edit_code) {
                        location.href = '/__code' + location.pathname + '/' + val;
                    }
                    if (is_edit_md) {
                        location.href = '/__md' + location.pathname + '/' + val;
                    }
                } else {
                    if (js.msg) {
                        alert(js.msg);
                    }
                }
            })
            .catch((err) => {
                console.error(err);
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
            let file_modal = document.getElementById('make_file_modal');

            file_modal.addEventListener('shown.bs.modal', () => {
                const inp = $('#make_file_input');
                if (inp) {
                    inp.focus();
                }
            });

            const mkfileModal = new bootstrap.Modal(file_modal, {});
            mkfileModal.show();
        });
    });

    // --------------------------------------------------------------------------------------------------------------------------------------

    const make_new_folder = async (ev) => {
        let val = $(ev.target.parentNode, 'input[type=text]').value;

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
                } else {
                    if (js.msg) {
                        alert(js.msg);
                    }
                }
            })
            .catch((err) => {
                console.error(err);
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
            let folder_modal = document.getElementById('make_folder_modal');

            folder_modal.addEventListener('shown.bs.modal', () => {
                const inp = $('#make_folder_input');
                if (inp) {
                    inp.focus();
                }
            });

            const mkfolderModal = new bootstrap.Modal(folder_modal, {});
            mkfolderModal.show();
        });
    });

    // --------------------------------------------------------------------------------------------------------------------------------------

    let search_submit = function (ev) {
        let val = $(ev.target.parentNode, 'input[type=text]').value;

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
            let modal = document.getElementById('search_modal');

            modal.addEventListener('shown.bs.modal', () => {
                const inp = $('#search_input');
                if (inp) {
                    inp.focus();
                }
            });

            const srchModal = new bootstrap.Modal(modal, {});
            srchModal.show();
        });
    });

    // --------------------------------------------------------------------------------------------------------------------------------------

    $$('a.rename').forEach((aa) => {
        aa.addEventListener('click', (ev) => {
            let el = ev.target;

            if (el.tagName == 'I') {
                el = el.parentNode;
            }
            //console.log('el=', el);

            $('#set_rename_orig').value = el.dataset.name;
            $('#set_rename_input').value = el.dataset.name;

            let mod = document.getElementById('rename_modal');

            mod.addEventListener('shown.bs.modal', () => {
                const inp = $('#rename_input');
                if (inp) {
                    inp.focus();
                }
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
                } else {
                    if (js.msg) {
                        alert(js.msg);
                    }
                }
            })
            .catch((err) => {
                console.error(err);
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

    // --------------------------------------------------------------------------------------------------------------------------------------

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
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    if( getCookie('code') ){
        
        const set_code_input = $('#set_code_input');
        if (set_code_input) {
            set_code_input.value = getCookie('code');
        }
    }
    
    let change_code_cookie = () => {
        
        const set_code_input = $('#set_code_input');
        if (set_code_input) {
            
            setCookie('code', set_code_input.value, {});
            
            location.href = location.href;
        }
    };
    
    const set_code_button = $('#set_code_button');
    if (set_code_button) {
        set_code_button.addEventListener('click', (ev) => {
            change_code_cookie(ev);
        });
    }
    
    const set_code_input = $('#set_code_input');
    if (set_code_input) {
        set_code_input.addEventListener('keypress', (ev) => {
            if (ev.which == 13) {
                change_code_cookie(ev);
            }
        });
    }

    // --------------------------------------------------------------------------------------------------------------------------------------

    
    $$('#upload_file, #upload_file2').forEach((upl) => {
        upl.addEventListener('change', (ev) => {
            ev_target_files(ev.target.files);
        });
    });
    

    // --------------------------------------------------------------------------------------------------------------------------------------
})();

// --------------------------------------------------------------------------------------------------------------------------------------

const ev_target_files = async (files) => {
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
            formData.append('fileBlob', file);
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
};

// --------------------------------------------------------------------------------------------------------------------------------------




// --------------------------------------------------------------------------------------------------------------------------------------
