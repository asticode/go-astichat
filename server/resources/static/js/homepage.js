var homepage = {
    // Init
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
    }
};