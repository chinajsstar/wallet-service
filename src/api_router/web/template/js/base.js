$("#logout").click(function(event){
    event.preventDefault();
    del_cookie("name");
    window.location.href = "/login";
})

function del_cookie(name)
{
    document.cookie = name + '=; expires=Thu, 01 Jan 1970 00:00:01 GMT;path=/;';
}

function Base64Encode(str, encoding='utf-8') {
    var bytes = new (TextEncoder || TextEncoderLite)(encoding).encode(str);
    return base64js.fromByteArray(bytes);
}

$("form[data-type=formRegisterAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");

    var uclass = ""+$('#uclass').attr("value")

    var argv = '{'
        + '"user_class":' + uclass + ''
        + ',"level":0'
        +'}'
    $.post(action, argv, function(ret){
        if(ret.err != 0) {
            alert(ret.errmsg);
        } else {
            alert('注册成功，请牢记user_key:'+ret.value.message);
            location.href = $(target).attr("form-rediret");
        }
    },"json")
})

$("form[data-type=formAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");

    var user_key = ""+$('#user_key').attr("value")
    var argv = '{"user_key":"'+ user_key + '"}'
    $.post(action, argv, function(ret){
        if(ret.err != 0) {
            alert(ret.errmsg);
        } else {
            location.href = $(target).attr("form-rediret");
        }
    },"json")
})

$("form[data-type=formDevSettingAction]").submit(function(event){
    event.preventDefault();
    var target = event.target;
    var action = $(target).attr("action");

    var pubkey = ""+$('#pubkey').attr("value")
    var sourceip = ""+$('#sourceip').attr("value")
    var cburl = ""+$('#cburl').attr("value")

    var pk = Base64Encode(pubkey)

    var argv = '{'
        + '"user_key":"' + "" +'"'
        + ',"public_key":"' + pk +'"'
        + ',"source_ip":"' + sourceip +'"'
        + ',"callback_url":"' + cburl +'"'
        +'}'
    alert(argv)
    $.post(action, argv, function(ret){
        if(ret.err != 0) {
            alert(ret.errmsg);
        } else {
            alert("设置成功，返回主页");
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