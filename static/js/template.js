function CheckMACAddress(a) {
    var b = /^(([0-9a-fA-F]{2}([:-]|$){0,1})){6}$|([0-9a-fA-F]{4}([.]|$)){3}$/;
    return !!b.test(a)
}
var macAddrs = $("input.mac_valid");
jQuery(document).ready(function () {
    $(".phone").mask("+7 (999) 999-9999", {autoclear: !1}), $("input.mac_valid").each(function () {
    }), $("input.mac_valid").on("input", '[data-action="text"]', function () {
        var a = $(this), b = a.val();
        console.log(CheckMACAddress(b))
    }), $("#cancel").click(function () {
        return $(".item:last-child").remove(), $(".item:last-child").is("div") || $("#cancel").css("display", "none"), i--, !1
    })
});
var i = 2;
$("#addNew").click(function () {
    function a(a) {
        a.selectionStart && (a.setSelectionRange(10, 5), a.focus())
    }

    $("#cancel").css("display", "inline-block"), $("#sbm").attr("disabled", "disabled"), macAddrs = $("input.mac_valid"), $(".input_forms").append("<div class='item'><h2 class='title'>Устройство #" + i + "<h2><div class='form-group'><label for='inputMac' class='col-md-2 control-label'>Mac</label><div class='col-md-10'><input type='mac' name='mac" + i + "' onchange=if(CheckMACAddress(this.value)==false){$(this).parent().parent().addClass('has-error1');$(this).parent().parent().removeClass('has-success');}else{$(this).parent().parent().addClass('has-success');$(this).parent().parent().removeClass('has-error1');} class='form-control mac_valid mac'  placeholder='Mac' required><span class='help-block'>Введите Mac адрес в формате XX:XX:XX:XX:XX:XX, где ХХ - цифры или латинские буквы от A до F</span></div></div>                                                                                                                       <div class='form-group'><label for='inputName' class='col-md-2 control-label'>Ф.И.О.</label> <div class='col-md-10'><input type='name' name='user" + i + "' class='form-control name' id='inputPassword' placeholder='Ф.И.О' required></div></div>                                          <div class='form-group'><label for='inputPassword' class='col-md-2 control-label'>Номер телефона</label><div class='col-md-10'>          <input type='telnumber' class='form-control phone'  name='tel" + i + "' id='ttl" + i + "' placeholder='Номер телефона' required></div></div></div>"), console.log("added"), $("#ttl" + i).mask("+7 (999) 999-9999", {autoclear: !0}), $("input.phone, input.mac, input.name").focusin(function () {
        $(this).parent().parent().removeClass("has-error1")
    });
    var b = "#ttl" + i;
    return a(document.querySelector(b)), $("input.phone").focusout(function () {
        $("input.phone").each(function () {
            17 !== $("input.phone").val().length || parseInt($("input.phone").val().indexOf("_")) !== -1 ? ($(this).parent().parent().addClass("has-error1"), $(this).parent().parent().removeClass("has-success")) : ($(this).parent().parent().addClass("has-success"), $(this).parent().parent().removeClass("has-error1"))
        })
    }), $("input.mac").focusout(function () {
        $("input.mac").each(function () {
            0 == CheckMACAddress($(this).val()) ? ($(this).parent().parent().addClass("has-error1"), $(this).parent().parent().removeClass("has-success")) : ($(this).parent().parent().addClass("has-success"), $(this).parent().parent().removeClass("has-error1"))
        })
    }), $("input.name").focusout(function () {
        $("input.name").each(function () {
            "" == $(this).val() ? ($(this).parent().parent().addClass("has-error1"), $(this).parent().parent().removeClass("has-success")) : ($(this).parent().parent().addClass("has-success"), $(this).parent().parent().removeClass("has-error1"))
        })
    }), $("input.phone, input.mac, input.name").keyup(function () {
        var a = !0, b = !0, c = !0;
        $("input.phone").each(function () {
            if (17 !== $(this).val().length || parseInt($(this).val().indexOf("_")) !== -1)return $("#sbm").attr("disabled", "disabled"), console.log($(this).val()), console.log("pE" + i), a = !1, !1
        }), $("input.mac").each(function () {
            if (0 == CheckMACAddress($(this).val()))return $("#sbm").attr("disabled", "disabled"), console.log("mE" + i), c = !1, !1
        }), $("input.name").each(function () {
            if ("" == $(this).val())return $("#sbm").attr("disabled", "disabled"), console.log("nE" + i), b = !1, !1
        }), console.log(b, c, a), b && c && a ? $("#sbm").removeAttr("disabled") : $("#sbm").attr("disabled", "disabled")
    }), i++, !1
}), $("input.phone, input.mac, input.name").focusin(function () {
    $(this).parent().parent().removeClass("has-error1")
}), $("input.phone").focusout(function () {
    17 !== $("input.phone").val().length || parseInt($("input.phone").val().indexOf("_")) !== -1 ? ($(this).parent().parent().addClass("has-error1"), $(this).parent().parent().removeClass("has-success")) : ($(this).parent().parent().addClass("has-success"), $(this).parent().parent().removeClass("has-error1"))
}), $("input.mac").focusout(function () {
    0 == CheckMACAddress($("input.mac").val()) ? ($(this).parent().parent().addClass("has-error1"), $(this).parent().parent().removeClass("has-success")) : ($(this).parent().parent().addClass("has-success"), $(this).parent().parent().removeClass("has-error1"))
}), $("input.name").focusout(function () {
    "" == $("input.name").val() ? ($(this).parent().parent().addClass("has-error1"), $(this).parent().parent().removeClass("has-success")) : ($(this).parent().parent().addClass("has-success"), $(this).parent().parent().removeClass("has-error1"))
}), $("input.phone, input.mac, input.name").keyup(function () {
    17 !== $("input.phone").val().length || parseInt($("input.phone").val().indexOf("_")) !== -1 || "" == $("input.name").val() || 0 == CheckMACAddress($("input.mac").val()) ? $("#sbm").attr("disabled", "disabled") : $("#sbm").removeAttr("disabled")
}), jQuery(document).ready(function () {
    function a(a) {
        a.selectionStart && (a.setSelectionRange(10, 5), a.focus())
    }

    jQuery("#info_win7, #info_win8, #info_mac, #info_android, #info_ios").hide(), a(document.querySelector("#inputPassword")), jQuery("a#tutorial").click(function () {
        return jQuery("#black-block").css("display", "block"), jQuery("#window").css("display", "block"), jQuery("#blurr").addClass("blurt"), jQuery("#win7").children("a").click(), jQuery("#w1").css("top", jQuery(window).scrollTop() + 214), !1
    }), jQuery("#black-block").click(function () {
        jQuery("#black-block").css("display", "none"), jQuery("#window").css("display", "none"), jQuery("body").removeClass("nooverflow"), jQuery("html").removeClass("nooverflow"), jQuery("#blurr").removeClass("blurt")
    }), jQuery(".closee").click(function () {
        jQuery("#black-block").css("display", "none"), jQuery("#window").css("display", "none"), jQuery("#blurr").removeClass("blurt")
    }), jQuery("#win7, #win8, #mac, #android, #ios").click(function () {
        jQuery("#info_win7, #info_win8, #info_mac, #info_android, #info_ios").hide(), jQuery("#info_" + jQuery(this).attr("id")).show()
    })
});