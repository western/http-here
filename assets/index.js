
// index.js


$(document).ready(function(){
    
    
    
    
    if( typeof(folderTree) != 'undefined' ){
    
        $('#folderTree').bstreeview({
            data: folderTree,
            expandIcon: 'bi bi-caret-down',
            collapseIcon: 'bi bi-caret-right',
            indent: 1.25,
            parentsMarginLeft: '1.25rem',
            openNodeLinkOnNewTab: true
        });
        
        $('div.list-group-item').click((ev) => {
            
            
            let path = $(ev.target).data('path');
            //console.log('path=', path);
            
            $('#move_folder_input').val(path);
        });
        
        $('#move_folder_button').click((ev) => {
            
            //console.log('button');
            
            
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
                
                //console.log('del=', checked);
                
                
                let formData = new FormData();
                
                formData.append('to', val);
                
                $.each(checked, function(indx, val){
                    //let i = indx+1;
                    //formData.append('name['+i+']', val);
                    formData.append('name', val);
                });
                
                //console.log('formData=', formData);
                
                
                
                
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
                });
                
                
            }
            
            
        });
    
    }
    
    
    
    
    
    
    $('#upload_file').on('change', function(ev){
        
        $('#signal').removeClass('visually-hidden');
        $(ev.target).prop('disabled', true);
        
        return ev_target_files(ev.target.files);
    });

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
    
    
    
    $('a.del').click((ev) => {
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
    
    
    $(':checkbox.head-chk').click((ev) => {
        
        //let checkboxes = $(':checkbox[name=fold]');
        //console.log('ev.target=', ev.target);
        
        let chk = $(ev.target).prop("checked") ? true : false;
        
        
        $(':checkbox[name=fold]').prop("checked", chk);
        $(':checkbox[name=file]').prop("checked", chk);
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
    
    })
    

});


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
            formData.append('fileMeta', JSON.stringify({
                lastModified: file.lastModified,
                lastModifiedDate: file.lastModifiedDate,
                name: file.name,
                size: file.size,
                type: file.type,
            }));
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
        xhr.open("POST", '/api/upload');
        xhr.send(formData);

    };

    submit();
    return;
}
