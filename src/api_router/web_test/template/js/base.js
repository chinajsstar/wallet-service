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

$("form[data-type=formAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");

    var name = ""+$('#name').attr("value")
    var ps = ""+$('#password').attr("value")
    var psmd5 = md5(ps)

    var argv = '{"user_name":"'+name + '","password":"' + psmd5 +'"}'
    $.post(action, argv, function(ret){
        if(ret.err != 0) {
            alert(ret.errmsg);
        } else {
            location.href = $(target).attr("form-rediret");
        }
    },"json")
})

$("form[data-type=testApiAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    //var action = $(target).attr("action");
    var path = "/wallet"+$('#method').attr("value")
    var argv = ""+$('#argv').attr("value")
    $.post(path, argv, function(ret){
        $('#err').html(ret.err)
        $('#errmsg').html(ret.errmsg)
        $('#message').html(ret.value.message)
    },"json")
})