//00:00:20:00:09:e1
function CheckMACAddress(MACAddress)
{
var RegExPattern = /^(([0-9a-fA-F]{2}([:-]|$){0,1})){6}$|([0-9a-fA-F]{4}([.]|$)){3}$/;

            if (!(RegExPattern.test(MACAddress))){
                console.log(MACAddress.length);
                return false;}
       
        else
return true;
}


jQuery(document).ready(function() {
    var i = 2;
    $("#addNew").click(function(){
        $(".input_forms").append("<div class='item'><h2 class='title'>User"+ i + "<h2><div class='form-group'><label for='inputMac' class='col-md-2 control-label'>Mac</label><div class='col-md-10'><input type='mac' name='mac"+i+"' class='form-control' id='inputEmail' placeholder='Mac' required></div></div>                                                                                                                       <div class='form-group'><label for='inputName' class='col-md-2 control-label'>Name</label> <div class='col-md-10'><input type='name' name='user"+i+"' class='form-control' id='inputPassword' placeholder='Ф.И.О' required></div></div>                                          <div class='form-group'><label for='inputPassword' class='col-md-2 control-label'>Phone Number</label><div class='col-md-10'>          <input type='telnumber' class='form-control' name='tel"+i+"' id='inputPassword' placeholder='Номер телефона' required></div></div></div>");
        i++;
        return false;

    });
    
    $("#inputEmail").change(function(){
        var q = this.value;
        console.log(q);
        console.log(CheckMACAddress(q));
    });
    
    $("#cancel").click(function(){
        $(".item:last-child").remove();
        return false;
    });
});