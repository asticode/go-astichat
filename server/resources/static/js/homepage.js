const h2Plain = 'a lightweight encrypted chat';
const h2Encrypted = 't afjgiesfuuj tpcrgsmfo ĝoek';
String.prototype.replaceAt=function(index, character) {
    return this.substr(0, index) + character + this.substr(index + 1, this.length - character.length);
};
var homepage = {
    // Init
    init: function() {
        // h2
        $("h2").text(h2Plain);
        var that = this;
        setTimeout(function () {
            that.handleH2();
            setInterval(that.handleH2.bind(that), 3000);
        }, 1000);

        // Has token checkbox
        var sel = $('#has_token');
        sel.change(function () {
            if($(this).is(":checked")) {
                $("#with_token").show();
                $("#without_token").hide();
            } else {
                $("#with_token").hide();
                $("#without_token").show();
            }
        });
        sel.prop("checked", false);
    },

    // H2
    handleH2: function() {
        var sel = $("h2");
        var text = h2Plain;
        if (sel.text() == h2Plain) {
            text = h2Encrypted;
        }
        this.type(text, sel);
    },
    type: function(message, sel) {
        (function writer(i){
            if(message.length <= i++){
                sel.text(message);
                return;
            }

            // Replace text
            if (sel.text().length < i) {
                sel.text(message.substring(0,i));
            } else {
                sel.text(sel.text().replaceAt(i-1, message[i-1]));
            }

            // Random timeout
            var rand = Math.floor(Math.random() * (50)) + 20;
            setTimeout(function(){writer(i);},rand);
        })(0)
    }
};