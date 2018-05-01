$("#logout").click(function(event){
    event.preventDefault();
    del_cookie("name");
    window.location.href = "/login";
})

function del_cookie(name)
{
    document.cookie = name + '=; expires=Thu, 01 Jan 1970 00:00:01 GMT;path=/;';
}

// function Base64Encode(str, encoding = 'utf-8') {
//     var bytes = new (TextEncoder || TextEncoderLite)(encoding).encode(str);
//     return base64js.fromByteArray(bytes);
// }

$("form[data-type=formOperationAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");
    $.post(action, $(target).serialize(), function(ret){
        if(ret.err != 0) {
            alert(ret.errmsg);
        } else {
            alert("操作成功：" + ret.value);
            location.href = $(target).attr("form-rediret");
        }
    },"json")
})

$("form[data-type=formUploadFileAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");

    var formData = new FormData();
    formData.append('uploadfile', document.getElementById('uploadfile').files[0]);
    $.ajax({
        type:"POST",
        url:action,
        data:formData,
        contentType:false,
        processData:false,
        success:function(data) {
            var ret = JSON.parse(data);
            if(ret.err != 0) {
                alert(ret.errmsg);
            } else {
                alert("操作成功：" + ret.value);
                location.href = $(target).attr("form-rediret");
            }
        }
    })
})