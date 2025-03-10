
// index.js


$(document).ready(function(){
    
    
    
    
    if( typeof(folderTree) != 'undefined' ){
        
        // --------------------------------------------------------------------------------------------------------------------------------------
        
        $('#group_move').click((ev) => {
            
            let checked = [];
            
            $(':checkbox[name=fold],:checkbox[name=file]').each((indx, el) => {
                
                
                if( $(el).prop('checked') ){
                    checked.push( $(el).val() );
                }
            });
            
            
            
            if( checked.length == 0 ){
                alert('You need select something')
            }
            
            if( checked.length > 0  ){
                
                
                new bootstrap.Offcanvas('#offcanvasMove').show()
            }
        
        });
        
        $('#folderTreeMove').bstreeview({
            data: folderTree,
            expandIcon: 'bi bi-caret-down',
            collapseIcon: 'bi bi-caret-right',
            indent: 1.25,
            parentsMarginLeft: '1.25rem',
            openNodeLinkOnNewTab: true
        });
        
        $('div.list-group-item').click((ev) => {
            
            
            let path = $(ev.target).data('path');
            
            
            $('#move_folder_input').val(path);
        });
        
        $('#move_folder_button').click((ev) => {
            
            
            
            
            let val = $('#move_folder_input').val();

            if ( val.length == 0 ){
                alert('Please choose folder');
                return;
            }
            
            
            let checked = [];
            
            $(':checkbox[name=fold],:checkbox[name=file]').each((indx, el) => {
                
                
                if( $(el).prop('checked') ){
                    checked.push( $(el).val() );
                }
            });
            
            
            
            if( checked.length == 0 ){
                alert('Please select some file')
            }
            
            if( checked.length > 0  ){
                
                
                
                let formData = new FormData();
                
                formData.append('to', val);
                
                $.each(checked, function(indx, val){
                    formData.append('name', val);
                });
                
                
                $.ajax({
                    url: '/api/move',
                    data: formData,
                    type: 'POST',
                    contentType: false,
                    processData: false,
                }).done(function( data ) {
                    
                    
                    if( data.code == 200 ){
                        location.href = location.href;
                    }else{
                        alert(data.msg);
                    }
                }).fail(function(data) {
                    if (data.responseJSON.msg) {
                        alert(data.responseJSON.msg);
                    }
                });
                
                
            }
            
            
        });
        
        // --------------------------------------------------------------------------------------------------------------------------------------
        
        // --------------------------------------------------------------------------------------------------------------------------------------
        
        $('#group_copy').click((ev) => {
            
            let checked = [];
            
            $(':checkbox[name=fold],:checkbox[name=file]').each((indx, el) => {
                
                
                if( $(el).prop('checked') ){
                    checked.push( $(el).val() );
                }
            });
            
            
            
            if( checked.length == 0 ){
                alert('You need select something')
            }
            
            if( checked.length > 0  ){
                
                
                new bootstrap.Offcanvas('#offcanvasCopy').show()
            }
        
        });
        
        
        $('#folderTreeCopy').bstreeview({
            data: folderTree,
            expandIcon: 'bi bi-caret-down',
            collapseIcon: 'bi bi-caret-right',
            indent: 1.25,
            parentsMarginLeft: '1.25rem',
            openNodeLinkOnNewTab: true
        });
        
        $('div.list-group-item').click((ev) => {
            
            
            let path = $(ev.target).data('path');
            
            
            $('#copy_folder_input').val(path);
        });
        
        $('#copy_folder_button').click((ev) => {
            
            
            
            
            let val = $('#copy_folder_input').val();

            if ( val.length == 0 ){
                alert('Please choose folder');
                return;
            }
            
            
            let checked = [];
            
            $(':checkbox[name=fold],:checkbox[name=file]').each((indx, el) => {
                
                
                if( $(el).prop('checked') ){
                    checked.push( $(el).val() );
                }
            });
            
            
            
            if( checked.length == 0 ){
                alert('Please select some file')
            }
            
            if( checked.length > 0  ){
                
                
                let formData = new FormData();
                
                formData.append('to', val);
                
                $.each(checked, function(indx, val){
                    formData.append('name', val);
                });
                
                
                
                $.ajax({
                    url: '/api/copy',
                    data: formData,
                    type: 'POST',
                    contentType: false,
                    processData: false,
                }).done(function( data ) {
                    
                    
                    if( data.code == 200 ){
                        location.href = location.href;
                    }else{
                        alert(data.msg);
                    }
                }).fail(function(data) {
                    if (data.responseJSON.msg) {
                        alert(data.responseJSON.msg);
                    }
                });
                
                
            }
            
            
        });
        
        
        
        
        
        
        // --------------------------------------------------------------------------------------------------------------------------------------
    
    }
    
    
    
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    

    let make_new_folder = function(ev){

        let val = $('input[type=text]', ev.target.parentNode).val();

        if ( val.length == 0 ){
            alert('Please fill folder name');
            return;
        }

        let formData = new FormData();

        formData.append('name', val);


        $.ajax({
            url: '/api/folder',
            data: formData,
            type: 'POST',
            contentType: false,
            processData: false,
        }).done(function( data ) {

            if( data.code == 200 ){
                location.href = location.href;
            }
        });

    };

    $('#make_folder_button').click(function(ev){

        make_new_folder(ev);
    });

    $('#make_folder_input').on('keypress', function(ev){

        if(ev.which == 13) {
            make_new_folder(ev);
        }
    });
    
    
    $('#make_folder_dlg, #make_folder_dlg2').click(function (ev) {
        
        let folder_modal = document.getElementById('make_folder_modal');
        
        folder_modal.addEventListener('shown.bs.modal', () => {
            $('#make_folder_input').focus()
        })
        
        const mkfolderModal = new bootstrap.Modal(folder_modal, {});
        mkfolderModal.show();
    });
    
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    
    let make_new_file = function (ev) {
        let val = $('input[type=text]', ev.target.parentNode).val();
        console.log('val=', val);

        if (val.length == 0) {
            alert('Please fill file name');
            return;
        }

        
        let formData = new FormData();
        formData.append('name', val);

        $.ajax({
            url: '/api/file/touch',
            data: formData,
            type: 'POST',
            contentType: false,
            processData: false,
        }).done(function (data) {
            if (data.code == 200) {
                //location.href = location.href;
                //location.href = '';
                
                /*
                let ext = path.parse(val).ext;
                ext = ext.replace(/\./g, '');
                ext = ext.toLowerCase();
                */
                
                let regx = /\.(.+?)$/.exec(val);
                
                //console.log('regx=', regx);
                
                if( !regx ){
                    location.href = location.href;
                    return;
                }
                
                let ext = regx.slice(1)[0];
                //console.log('ext=', ext);
                
                let is_edit_doc = ext.match(/^(rtf|doc|docx|odt)$/i);
                let is_edit_code = ext.match(/^(html|txt|js|css|md)$/i);
                
                if ( is_edit_doc ){
                    location.href = '/__doc' + location.pathname + '/' + val;
                }
                if ( is_edit_code ){
                    location.href = '/__code' + location.pathname + '/' + val;
                }
                
            }
        }).fail(function(data) {
            if (data.responseJSON.msg) {
                alert(data.responseJSON.msg);
            }
        });
        
    };

    $('#make_file_button').click(function (ev) {
        make_new_file(ev);
    });

    $('#make_file_input').on('keypress', function (ev) {
        if (ev.which == 13) {
            make_new_file(ev);
        }
    });
    
    
    $('#new_file_dlg, #new_file_dlg2').click(function (ev) {
        
        let file_modal = document.getElementById('make_file_modal');
        
        file_modal.addEventListener('shown.bs.modal', () => {
            $('#make_file_input').focus()
        })
        
        const mkfileModal = new bootstrap.Modal(file_modal, {});
        mkfileModal.show();
    });
    
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    let search_submit = function (ev) {
        let val = $('input[type=text]', ev.target.parentNode).val();
        console.log('val=', val);

        if (val.length == 0) {
            alert('Please fill search field');
            return;
        }

        
        location.href = '/__search/?s=' + val;
        
    };
    

    $('#search_button').click(function (ev) {
        
        search_submit(ev);
    });

    $('#search_input').on('keypress', function (ev) {
        if (ev.which == 13) {
            search_submit(ev);
        }
    });
    
    
    $('#search_dlg, #search_dlg2').click(function (ev) {
        
        
        
        let search_modal = document.getElementById('search_modal');
        
        search_modal.addEventListener('shown.bs.modal', () => {
            $('#search_input').focus()
        })
        
        const searchModal = new bootstrap.Modal(search_modal, {});
        searchModal.show();
    });
    
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    $(':checkbox.head-chk').click((ev) => {
        
        //let checkboxes = $(':checkbox[name=fold]');
        //console.log('ev.target=', ev.target);
        
        let chk = $(ev.target).prop("checked") ? true : false;
        
        
        $(':checkbox[name=fold]').prop("checked", chk);
        $(':checkbox[name=file]').prop("checked", chk);
    })
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    $('a.edit_doc').click((ev) => {
        let el = ev.target;
        
        
        
        if(el.tagName == 'I'){
            el = el.parentNode;
        }
        
        //console.log('a.edit', el);
        
        let full_path = $(el).data('full-path');
        if(full_path){
            location.href = '/__doc' + full_path;
        }
        
    });
    
    $('a.edit_code').click((ev) => {
        let el = ev.target;
        
        
        
        if(el.tagName == 'I'){
            el = el.parentNode;
        }
        
        //console.log('a.edit', el);
        
        let full_path = $(el).data('full-path');
        if(full_path){
            location.href = '/__code' + full_path;
        }
        
    });
    
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    $('a.rename').click((ev) => {
        let el = ev.target;
        
        //console.log('a.rename', el);
        
        if(el.tagName == 'I'){
            el = el.parentNode;
        }
        
        //console.log('a.rename', el);
        
        $('#set_rename_orig').val( $(el).data('name') );
        $('#set_rename_input').val( $(el).data('name') );

        
        
        
        let rename_modal = document.getElementById('rename_modal');
        
        rename_modal.addEventListener('shown.bs.modal', () => {
            $('#set_rename_input').focus()
        })
        
        const renameModal = new bootstrap.Modal(rename_modal, {});
        renameModal.show();
        
    });
    
    let rename_form_submit = () => {
        
        let orig = $('#set_rename_orig').val();
        let name = $('#set_rename_input').val();
        
        let formData = new FormData();
        formData.append('name', orig);
        formData.append('to', name);

        $.ajax({
            url: '/api/rename',
            data: formData,
            type: 'POST',
            contentType: false,
            processData: false,
        }).done(function (data) {
            if (data.code == 200) {
                location.href = location.href;
            }
        }).fail(function(data) {
            if (data.responseJSON.msg) {
                alert(data.responseJSON.msg);
            }
        });
        
    };
    
    $('#set_rename_button').click((ev) => {
        
        rename_form_submit();
    });

    $('#set_rename_input').on('keypress', (ev) => {
        if (ev.which == 13) {
            
            rename_form_submit();
        }
    });
    
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    $('a.del, button.del').click((ev) => {
        let el = ev.target;
        
        //console.log('el=', el);
        
        if(el.tagName == "I"){
            el = el.parentNode;
            //console.log('el=', $(el).data('name'));
            
            
            if( confirm('Delete "'+$(el).data('name')+'"?') ){
                
                let formData = new FormData();

                formData.append('name', $(el).data('name'));
                //formData.append('name[1]', $(el).data('name'));
                //formData.append('name[2]', $(el).data('name')+'_2');
                
                $.ajax({
                    url: '/api/delete',
                    data: formData,
                    type: 'POST',
                    contentType: false,
                    processData: false,
                }).done(function( data ) {
                    
                    
                    if( data.code == 200 ){
                        location.href = location.href;
                    }else{
                        alert(data.msg);
                    }
                });
            }
        }
    })
    
    
    
    
    
    $('#group_del').click((ev) => {
        
        let checked = [];
        
        $(':checkbox[name=fold],:checkbox[name=file]').each((indx, el) => {
            
            
            if( $(el).prop('checked') ){
                checked.push( $(el).val() );
            }
        });
        
        //console.log('checked=', checked);
        
        if( checked.length == 0 ){
            alert('You need select something')
        }
        
        if( checked.length > 0 && confirm('You really want to del this group?') ){
            
            //console.log('del=', checked);
            
            
            let formData = new FormData();
            
            $.each(checked, function(indx, val){
                //let i = indx+1;
                //formData.append('name['+i+']', val);
                formData.append('name', val);
            });
            
            //console.log('formData=', formData);
            
            
            
            
            $.ajax({
                url: '/api/delete',
                data: formData,
                type: 'POST',
                contentType: false,
                processData: false,
            }).done(function( data ) {
                
                
                if( data.code == 200 ){
                    location.href = location.href;
                }else{
                    alert(data.msg);
                }
            });
            
            
        }
        
    })
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    $('#group_zip').click((ev) => {
        
        let checked = [];
        
        $(':checkbox[name=fold],:checkbox[name=file]').each((indx, el) => {
            
            
            if( $(el).prop('checked') ){
                checked.push( $(el).val() );
            }
        });
        
        
        
        if( checked.length == 0 ){
            alert('You need select something')
        }
        
        if( checked.length > 0  ){
            
            $(ev.target).prop('disabled', true);
            
            let formData = new FormData();
            
            $.each(checked, function(indx, val){
                //let i = indx+1;
                //formData.append('name['+i+']', val);
                formData.append('name', val);
            });
            
            
            
            
            
            
            $.ajax({
                url: '/api/zip',
                data: formData,
                type: 'POST',
                contentType: false,
                processData: false,
            }).done(function( data ) {
                
                
                
                if( data.code == 200 ){
                    
                    $(ev.target).prop('disabled', false);
                    window.location.href = data.file;
                }
                
                
                
            });
            
            
        }
        
    })
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    
    
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    
    $('#set_code_button').click(function(ev){
        
        let val = $('#set_code_input').val();
        //document.cookie = "code="+val;
        setCookie('code', val, {});
        
        location.href = location.href;
    });

    $('#set_code_input').on('keypress', function(ev){

        if(ev.which == 13) {
            let val = $('#set_code_input').val();
            //document.cookie = "code="+val;
            setCookie('code', val, {});
            
            location.href = location.href;
        }
    });
    
    if( getCookie('code') ){
        $('#set_code_input').val( getCookie('code') )
    }
    
    // --------------------------------------------------------------------------------------------------------------------------------------
    
    
    
    
    
    
    
    $('#upload_file, #upload_file2').on('change', function(ev){
        
        $('#signal').removeClass('visually-hidden');
        $(ev.target).prop('disabled', true);
        
        return ev_target_files(ev.target.files);
    });
    
    

});


// --------------------------------------------------------------------------------------------------------------------------------------

function ev_target_files(files){
    


    if ( files.length > config.files_count_max ){

        alert(`Count of files is more than ${config.files_count_max}.`);
        location.href = location.href;
        return;
    }

    let formData = new FormData();
    Array.prototype.forEach.call(files, function(file) {


        if ( file.size > config.fieldSize_max ){
            alert( 'File "' + file.name + `" size is overload "${config.fieldSize_max_human}"`);
        }else{

            formData.append('fileBlob', file);
            /*
            formData.append('fileMeta', JSON.stringify({
                lastModified: file.lastModified,
                lastModifiedDate: file.lastModifiedDate,
                name: file.name,
                size: file.size,
                type: file.type,
            }));*/
        }
    });

    let submit = async function() {

        /*
        let response = await fetch('/api/upload', {
            method: 'POST',
            body: formData
        });

        let result = await response.json();
        if(result.code == 200){
            location.href = location.href;
        }
        */


        $('#progress').attr('max', 100);
        $('#progress').attr('value', 0);
        $('#progress').show();

        let xhr = new XMLHttpRequest();

        /*
        xhr.addEventListener("error", function(e){
            console.log(e, `${e.type}: ${e.loaded} bytes transferred`);
        });
        xhr.addEventListener("abort", function(e){
            console.log(e, `${e.type}: ${e.loaded} bytes transferred`);
        });
        */

        xhr.upload.addEventListener('progress', function(ev, th){
            if (ev.lengthComputable) {

                //let percentComplete = (ev.loaded / ev.total) * 100;
                //console.log('percentComplete=', percentComplete);

                $('#progress').attr('max', ev.total);
                $('#progress').attr('value', ev.loaded);
            }
        }, false);

        xhr.onreadystatechange = function (ev) {
            if (xhr.readyState == 4) {
                location.href = location.href;
            }
        };
        xhr.open("POST", '/api/file/upload');
        xhr.send(formData);

    };

    submit();
    return;
}



// --------------------------------------------------------------------------------------------------------------------------------------


function getCookie(name) {
  let matches = document.cookie.match(new RegExp(
    "(?:^|; )" + name.replace(/([\.$?*|{}\(\)\[\]\\\/\+^])/g, '\\$1') + "=([^;]*)"
  ));
  return matches ? decodeURIComponent(matches[1]) : undefined;
}

function setCookie(name, value, options = {}) {

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

    //console.log('updatedCookie=', updatedCookie);
    document.cookie = updatedCookie;
}

// setCookie('user', 'John', {secure: true, 'max-age': 3600});


