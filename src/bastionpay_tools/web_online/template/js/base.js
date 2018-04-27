$("#logout").click(function(event){
    event.preventDefault();
    del_cookie("name");
    window.location.href = "/login";
})

function del_cookie(name)
{
    document.cookie = name + '=; expires=Thu, 01 Jan 1970 00:00:01 GMT;path=/;';
}

function Base64Encode(str, encoding = 'utf-8') {
    var bytes = new (TextEncoder || TextEncoderLite)(encoding).encode(str);        
    return base64js.fromByteArray(bytes);
}

$("form[data-type=formBuildtxAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");
    $.post(action, $(target).serialize(), function(ret){
        if(ret.err != 0) {
            alert(ret.errmsg);
        } else {
            alert("操作成功，请记录文件名：" + ret.errmsg);
            location.href = $(target).attr("form-rediret");
        }
    },"json")
})

$("form[data-type=formUploadSignedtxAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");
    $.post(action, $(target).serialize(), function(ret){
        if(ret.err != 0) {
            alert(ret.errmsg);
        } else {
            alert("操作成功：" + ret.errmsg);
            location.href = $(target).attr("form-rediret");
        }
    },"json")
})

$("form[data-type=formUploadSignedtxAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");

    var fd = new FormData();
    fd.append("upload", 1);
    fd.append("uploadfile", $("#uploadfile").get(0).files[0]);
    $.ajax({
        url: action,
        type: "POST",
        processData: false,
        contentType: false,
        data: fd,
        success: function(ret) {
            if(ret.err != 0) {
                alert(ret.errmsg);
            } else {
                alert("操作成功：" + ret.errmsg);
                location.href = $(target).attr("form-rediret");
            }
        }
    }, "json")

    // $.post(action, $(target).serialize(), function(ret){
    //     if(ret.err != 0) {
    //         alert(ret.errmsg);
    //     } else {
    //         alert("操作成功：" + ret.errmsg);
    //         location.href = $(target).attr("form-rediret");
    //     }
    // },"json")
})

$("form[data-type=formSendSignedtxAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");

    $.post(action, $(target).serialize(), function(ret){
        if(ret.err != 0) {
            alert(ret.errmsg);
        } else {
            alert("操作成功：" + ret.errmsg);
            location.href = $(target).attr("form-rediret");
        }
    },"json")
})