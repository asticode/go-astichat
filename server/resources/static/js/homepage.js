var homepage = {
    init: function() {
        // Upgrade checkbox
        var sel = $('#is_upgrade');
        sel.change(function () {
            if($(this).is(":checked")) {
                $("#with_upgrade").show();
                $("button").text("Upgrade");
            } else {
                $("#with_upgrade").hide();
                $("button").text("Download");
            }
        });
        sel.prop("checked", false);

        // Errors
        $("button").click(function() {
            $('.form-error').hide();
        });
        if (getQueryParam('error') != '') {
            $('.form-error').show();
            $('.form-error .alert').text(getQueryParam('error'));
        }
    }
};

var getQueryParam = function(name) {
    name = name.replace(/[\[]/, '\\[').replace(/[\]]/, '\\]');
    var regex = new RegExp('[\\?&]' + name + '=([^&#]*)');
    var results = regex.exec(location.search);
    return results === null ? '' : decodeURIComponent(results[1].replace(/\+/g, ' '));
};