function CheckMACAddress(MACAddress) {
    var RegExPattern = /^(([0-9a-fA-F]{2}([:-]|$){0,1})){6}$|([0-9a-fA-F]{4}([.]|$)){3}$/;
    return RegExPattern.test(MACAddress);
}

function CheckName(Name) {
    var RegExPattern = /[^а-яА-Яa-zA-Z \-]+$/;
    newName = Name.replace(RegExPattern, "");
    test =  newName === Name;
    return true
}

(function($, root, undefined) {
"use strict"
var macAddrs = $("input.mac_valid");
$(document).ready(function () {
    $('.phone').mask("+7 (999) 999-9999", {autoclear: false});

    $("input.mac_valid").on('input', '[data-action="text"]', function () {
        var $item = $(this),
            value = $item.val();
    });

    $("#cancel").click(function () {
        $(".item:last-child").remove();
        if (!$(".item:last-child").is("div")) {
            $("#cancel").css("display", "none");
        }
        i--;
	$('#sbm').removeAttr('disabled');
        return false;
    });
    if ($('.search')[0]) {
        $('.search').click(function () {
            if($(this).hasClass("glyphicon-ok")){
                $(this).removeClass("glyphicon-ok");
                $(this).addClass("glyphicon-remove");
            }
            else {
                console.log("1337");
                $(this).addClass("glyphicon-ok");
                $(this).removeClass("glyphicon-remove");
            }
        });
    }
});

var i = 2;
$("#addNew").click(function () {
    $("#cancel").css("display", "inline-block");
    $('#sbm').attr('disabled', 'disabled');
    macAddrs = $("input.mac_valid");
    $(".input_forms").append("<div class='item'><h2 class='title'>Устройство #" + i + "<h2><div class='form-group'><label for='inputMac' class='col-md-2 control-label'>Mac</label><div class='col-md-10'><input type='mac' name='mac" + i + "' onchange=if(CheckMACAddress(this.value)==false){$(this).parent().parent().addClass('has-error1');$(this).parent().parent().removeClass('has-success');}else{$(this).parent().parent().addClass('has-success');$(this).parent().parent().removeClass('has-error1');} class='form-control mac_valid mac'  placeholder='Mac' required><span class='help-block'>Введите Mac адрес в формате XX:XX:XX:XX:XX:XX, где ХХ - цифры или латинские буквы от A до F</span></div></div>                                                                                                                       <div class='form-group'><label for='inputName' class='col-md-2 control-label'>Ф.И.О.</label> <div class='col-md-10'><input type='name' name='user" + i + "' class='form-control name' id='inputPassword' placeholder='Ф.И.О' required></div></div>                                          <div class='form-group'><label for='inputPassword' class='col-md-2 control-label'>Номер телефона</label><div class='col-md-10'>          <input type='telnumber' class='form-control phone'  name='tel" + i + "' id='ttl" + i + "' placeholder='Номер телефона' required></div></div></div>");

    $('#ttl' + i).mask("+7 (999) 999-9999", {autoclear: true});
    $("input.phone, input.mac, input.name").focusin(function () {
        $(this).parent().parent().removeClass("has-error1");
    });
    function moveCaretToStart(inputObject) {
        if (inputObject.selectionStart) {
            inputObject.setSelectionRange(10, 5);
            inputObject.focus();
        }
    }

    var q = "#ttl" + i;
    moveCaretToStart(document.querySelector(q));
    $("input.phone").focusout(function () {
        $("input.phone").each(function () {
            if ($("input.phone").val().length !== 17 || parseInt($("input.phone").val().indexOf("_")) !== -1 ) {
                $(this).parent().parent().addClass('has-error1');
                $(this).parent().parent().removeClass('has-success');
            } else {
                $(this).parent().parent().addClass('has-success');
                $(this).parent().parent().removeClass('has-error1');
            }
        })
    });
    $("input.mac").focusout(function () {
        $("input.mac").each(function () {
            if (!CheckMACAddress($(this).val())) {
                $(this).parent().parent().addClass('has-error1');
                $(this).parent().parent().removeClass('has-success');
            } else {
                $(this).parent().parent().addClass('has-success');
                $(this).parent().parent().removeClass('has-error1');
            }
        })
    });
    $("input.name").focusout(function () {
        $("input.name").each(function () {
            if (!CheckName($(this).val())) {
                $(this).parent().parent().addClass('has-error1');
                $(this).parent().parent().removeClass('has-success');
            } else {
                $(this).parent().parent().addClass('has-success');
                $(this).parent().parent().removeClass('has-error1');
            }
        })
    });

    $("input.phone, input.mac, input.name").keyup(function () {
        var p = true, n = true, m = true;
        $("input.phone").each(function () {
            if ($(this).val().length !== 17 || parseInt($(this).val().indexOf("_")) !== -1) {
                $('#sbm').attr('disabled', 'disabled');
                console.log($(this).val());
                console.log("pE" + i);
                p = false;
                return false;
            }
        });

        $("input.mac").each(function () {
            if (!CheckMACAddress($(this).val())) {
                $('#sbm').attr('disabled', 'disabled');
                console.log("mE" + i);
                m = false;
                return false;
            }
        });
        $("input.name").each(function () {
            if ($(this).val() == "") {
                $('#sbm').attr('disabled', 'disabled');
                console.log("nE" + i);
                n = false;
                return false;
            }
        });
        if (n && m && p) {
            $('#sbm').removeAttr('disabled');
        }
        else {
            $('#sbm').attr('disabled', 'disabled');
        }
    });

    i++;
    return false;
});


$("input.phone, input.mac, input.name").focusin(function () {
    $(this).parent().parent().removeClass("has-error1");
});

$("input.phone").focusout(function () {
    if ($("input.phone").val().length !== 17 || parseInt($("input.phone").val().indexOf("_")) !== -1 ) {
        $(this).parent().parent().addClass('has-error1');
        $(this).parent().parent().removeClass('has-success');
    } else {
        $(this).parent().parent().addClass('has-success');
        $(this).parent().parent().removeClass('has-error1');
    }
});

$("input.mac").focusout(function () {
    if (!CheckMACAddress($("input.mac").val())) {
        $(this).parent().parent().addClass('has-error1');
        $(this).parent().parent().removeClass('has-success');
    }
    else {
        $(this).parent().parent().addClass('has-success');
        $(this).parent().parent().removeClass('has-error1');
    }
});
$("input.name").focusout(function () {
    if (($("input.name").val() == "") || (!CheckName($(this).val()))) {
        $(this).parent().parent().addClass('has-error1');
        $(this).parent().parent().removeClass('has-success');
    }
    else {
        $(this).parent().parent().addClass('has-success');
        $(this).parent().parent().removeClass('has-error1');
    }
});

$("input.phone, input.mac, input.name").keyup(function () {
    if ($("input.phone").val().length !== 17 || parseInt($("input.phone").val().indexOf("_")) !== -1 || $("input.name").val() == "" || !CheckMACAddress($("input.mac").val()) || !CheckName($("input.name").val())) {
        console.log("dis");
        $('#sbm').attr('disabled', 'disabled');
    }
    else {
        console.log("en");
        $('#sbm').removeAttr('disabled');
    }
});

$(document).ready(function () {
    $("#info_win7, #info_win8, #info_mac, #info_android, #info_ios").hide();
    function moveCaretToStart(inputObject) {
        if (inputObject.selectionStart) {
            inputObject.setSelectionRange(10, 5);
            inputObject.focus();
        }
    }
    if($('.phone')[0]) {
        moveCaretToStart(document.querySelector('.phone'));
    }
    $("a#tutorial").click(function () {
        $("#black-block").css('display', 'block');
        $("#window").css('display', 'block');
        $("#blurr").addClass("blurt");
        $("#win7").children('a').click();
        $("#w1").css("top", $(window).scrollTop() + 214);
        return false;
    });
    $("a#guide").click(function () {
        $("#black-block").css('display', 'block');
        $("#window1").css('display', 'block');
        $("#blurr").addClass("blurt");
        //$("#win7").children('a').click();
        //$("#w1").css("top", $(window).scrollTop() + 214);
        return false;
    });

    $("#black-block").click(function () {
        $("#black-block").css('display', 'none');
        $("#window").css('display', 'none');
        $("#window1").css('display', 'none');
        $("body").removeClass("nooverflow");
        $("html").removeClass("nooverflow");
        $("#blurr").removeClass("blurt");

    });
    $(".closee").click(function () {
        $("#black-block").css('display', 'none');
        $("#window").css('display', 'none');
        $("#window1").css('display', 'none');
        $("#blurr").removeClass("blurt");
    });

    $("#win7, #win8, #mac, #android, #ios").click(function () {
        $("#info_win7, #info_win8, #info_mac, #info_android, #info_ios").hide();
        $("#info_" + $(this).attr('id')).show();
    });
});
})($, this);