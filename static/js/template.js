function CheckMACAddress(MACAddress) {
    var RegExPattern = /^(([0-9a-fA-F]{2}([:-]|$){0,1})){6}$|([0-9a-fA-F]{4}([.]|$)){3}$/;
    return RegExPattern.test(MACAddress);
}

/**
 * @return {boolean}
 */
function CheckName(Name) {
    var RegExPattern = /^[a-zA-Zа-яА-ЯёЁ0-9'][a-zA-Z-а-яА-ЯёЁ0-9' ]+[a-zA-Zа-яА-ЯёЁ0-9']?$/;
    // newName = Name.replace(RegExPattern, "");
    // test = newName === Name;
    return RegExPattern.test(Name);
}

(function ($, root, undefined) {
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
            var cont = true;
            $('input[type="mac"]').each(
                function () {
                    if (!CheckMACAddress($(this).val())) {
                        cont = false;
                    }
                }
            );

            if (cont) {
                $('#sbm').removeAttr('disabled');
            }
            return false;
        });
        if ($('.search')[0]) {
            $('.search').click(function () {
                if ($(this).hasClass("glyphicon-ok")) {
                    $(this).removeClass("glyphicon-ok");
                    $(this).addClass("glyphicon-remove");
                }
                else {
                    // console.log("1337");
                    $(this).addClass("glyphicon-ok");
                    $(this).removeClass("glyphicon-remove");
                }
            });
        }
    });
      $("#ip_phone").change(function(){
        if(this.checked){
            $('#ip_numbers').show();
            $('#addNumb_ip').show();
            $('#delNumb_ip').show();
        }else{
            $('#ip_numbers').hide();
            $('#addNumb_ip').hide();
            $('#delNumb_ip').hide();
        }

    });
    var q = 2;
    $("#addNumb").click(function () {
        $("#delNumb").css("display", "inline-block");
        $("#cancel").css("display", "inline-block");
        $('#sbm').attr('disabled', 'disabled');
        macAddrs = $("input.mac_valid");
        $("#number").append("          <div class='numb_per col-12'>      <h2 class='title col-12'>Абонент #"+ q +"</h2>\n" +
            "                <div class=\"numbers col-12\">\n" +
            "                    <fieldset class=\"form-group\">\n" +
            "                        <label for='TelNumb'"+ q +" class=\"bmd-label-static\">Желаемый номер</label>\n" +
            "                        <input type=\"text\" name='num"+q+"' class=\"form-control\" placeholder=\"Желаемый номер\" id='TelNumb'\"+ q +\">\n" +
            "                        <span class=\"bmd-help\">Желаемый номер</span>\n" +
            "                    </fieldset>\n" +
            "                    <div class=\"form-group\">\n" +
            "                        <label for='typePhone1'"+ q +" class=\"bmd-label-floating\">Тип телефона</label>\n" +
            "                        <select class=\"form-control\" name='typePhone1'"+ q +"  id='typePhone1'"+ q +">\n" +
            "                            <option value='1'>Внутренний</option>\n" +
            "                            <option value='2'>Городской</option>\n" +
            "                            <option value='3'>Меж.городской</option>\n" +
            "                            <option value='4'>Международный</option>\n" +
            "                        </select>\n" +
            "                    </div>\n" +
            "                    <div class=\"form-group col-12 row\">\n" +
            "                        <fieldset class=\"form-group\">\n" +
            "                            <label for='room'"+q+" class=\"bmd-label-static\">Кабинет</label>\n" +
            "                            <input type=\"text\" name='room'"+q+" class=\"form-control\" placeholder=\"Кабинет\" id='room'"+q+" >\n" +
            "                            <span class=\"bmd-help\">Кабинет</span>\n" +
            "                        </fieldset>\n" +
            "                        <fieldset class=\"form-group offset-1\">\n" +
            "                            <label for='build'"+q+" class=\"bmd-label-static\">Корпус</label>\n" +
            "                            <input type=\"text\" name='build'"+q+" class=\"form-control\" placeholder=\"Корпус\" id='build'"+q+">\n" +
            "                            <span class=\"bmd-help\">Корпус</span>\n" +
            "                        </fieldset>\n" +
            "                    </div>\n" +
            "                </div></div>");
        q++;
    });
    $("#delNumb").click(function () {
        $(".numb_per:last-child").remove();
        if (!$(".numb_per:last-child").is("div")) {
            $("#delNumb").css("display", "none");
        }
        q--;
    });
    var w = 2;
    $("#addNumb_ip").click(function () {
        $("#delNumb_ip").css("display", "inline-block");
        $("#cancel").css("display", "inline-block");
        $('#sbm').attr('disabled', 'disabled');
        macAddrs = $("input.mac_valid");
        $("#ip_numbers").append("<div class='numb_per_ip col-12'>                <fieldset class=\"form-group col-12\">\n" +
            "                    <label for='TelNumb_ip_"+w+"' class=\"bmd-label-static\">Номер</label>\n" +
            "                    <input type=\"text\" name='TelNumb_ip_"+w+"' class=\"form-control\" placeholder=\"Номер\" id='TelNumb_ip_"+w+"'>\n" +
            "                    <span class=\"bmd-help\">Номер</span>\n" +
            "                </fieldset>\n" +
            "                <div class=\"form-group col-12 row\">\n" +
            "                    <fieldset class=\"form-group col-2\">\n" +
            "                        <label for='room_ip_"+w+"' class=\"bmd-label-static\">Кабинет</label>\n" +
            "                        <input type=\"text\" name='room_ip_"+w+"' class=\"form-control\" placeholder=\"Кабинет\" id='room_ip_"+w+"' >\n" +
            "                        <span class=\"bmd-help\">Кабинет</span>\n" +
            "                    </fieldset>\n" +
            "                    <fieldset class=\"form-group offset-1 col-2\">\n" +
            "                        <label for='build_ip_"+w+"' class=\"bmd-label-static\">Корпус</label>\n" +
            "                        <input type=\"text\" name='build_ip_"+w+"' class=\"form-control\" placeholder=\"Корпус\" id='build_ip_"+w+"'>\n" +
            "                        <span class=\"bmd-help\">Корпус</span>\n" +
            "                    </fieldset>\n" +
            "                </div></div>");
        w++;
    });
    $("#delNumb_ip").click(function () {
        $(".numb_per_ip:last-child").remove();
        if (!$(".numb_per_ip:last-child").is("div")) {
            $("#delNumb_ip").css("display", "none");
        }
        w--;
    });
    var i = 2;
    $("#addNew").click(function () {
        $("#cancel").css("display", "inline-block");
        $('#sbm').attr('disabled', 'disabled');
        macAddrs = $("input.mac_valid");
        $(".input_forms").append("<div class='item'>" +
            "<h2 class='title'>Устройство #" + i + "</h2>" +
            "<div class='form-group bmd-form-group'>" +
                "<label for='inputMac' class='col-md-2 control-label bmd-label-static'>Mac</label>" +
                "<div class='col-md-10'>" +
                    "<input type='mac' name='mac" + i + "' onchange=if(CheckMACAddress(this.value)==false){$(this).parent().parent().addClass('has-error1');$(this).parent().parent().removeClass('has-success');}else{$(this).parent().parent().addClass('has-success');$(this).parent().parent().removeClass('has-error1');} class='form-control mac_valid mac'  placeholder='Mac' required>" +
                    "<span class='help-block bmd-help'>Введите Mac адрес в формате XX:XX:XX:XX:XX:XX, где ХХ - цифры или латинские буквы от A до F</span>" +
                "</div>" +
            "</div>" +
            "<div class='form-group bmd-form-group'>" +
                "<label for='inputName' class='col-md-2 control-label bmd-label-static'>Ф.И.О.</label>" +
            "<div class='col-md-10'>" +
                        "<input type='name' onchange=if(CheckName(this.value)==false){$(this).parent().parent().addClass('has-error1');$(this).parent().parent().removeClass('has-success');}else{$(this).parent().parent().addClass('has-success');$(this).parent().parent().removeClass('has-error1');} name='user" + i + "' class='form-control name' id='inputPassword' placeholder='Ф.И.О' required>" +
                        "<span class='help-block bmd-help'>Для служебных устройств - наименование подразделения</span>" +
                    "</div>" +
            "</div>" +
            "<div class='form-group bmd-form-group'>" +
                "<label for='inputPassword' class='col-md-2 control-label bmd-label-static'>Номер телефона</label>" +
                "<div class='col-md-10'>" +
                    "<input type='telnumber' class='form-control phone'  name='tel" + i + "' id='ttl" + i + "' placeholder='Номер телефона' required>" +
                    "<span class='help-block bmd-help'>Для служебных устройств - служебный номер подразделения</span>" +
                "</div>" +
            "</div>" +
            "</div>");

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
        $("input.phone, input.mac, input.name").focusin(function () {
            $(this).parent().parent().addClass('is-focused');
        });
        $("input.phone, input.mac, input.name").focusout(function () {
            $(this).parent().parent().removeClass('is-focused');
        });
        $("input.phone").focusout(function () {
            $("input.phone").each(function () {
                if ($("input.phone").val().length !== 17 || parseInt($("input.phone").val().indexOf("_")) !== -1) {
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
                    // console.log($(this).val());
                    // console.log("pE" + i);
                    p = false;
                    return false;
                }
            });

            $("input.mac").each(function () {
                if (!CheckMACAddress($(this).val())) {
                    $('#sbm').attr('disabled', 'disabled');
                    // console.log("mE" + i);
                    m = false;
                    return false;
                }
            });
            $("input.name").each(function () {
                if ($(this).val() == "") {
                    $('#sbm').attr('disabled', 'disabled');
                    // console.log("nE" + i);
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
        if ($("input.phone").val().length !== 17 || parseInt($("input.phone").val().indexOf("_")) !== -1) {
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
            // console.log("dis");
            $('#sbm').attr('disabled', 'disabled');
        }
        else {
            // console.log("en");
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

        if ($('.phone')[0]) {
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