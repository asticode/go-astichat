var homepage = {
    init: function() {
        $('#has_public_key').change(function () {
            var sel = $('#public_key');
            if($(this).is(":checked")) {
                sel.show();
            } else {
                sel.hide();
            }
        });
        $('#has_public_key').prop("checked", false);
    }
};