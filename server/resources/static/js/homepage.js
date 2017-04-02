var homepage = {
    init: function() {
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

        // Errors
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