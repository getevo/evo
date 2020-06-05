var UPGRADE_DEFAULT_DATE = {
    locale: {
        format: "YYYY/MM/DD",
    },
    showDropdowns:true,
    isInvalidDate:false
}
var Upgrade = function (parent) {
    var numeric = $(parent).find("input[type=number]")
    if(numeric.length > 0) {
        require(['plugins/jquery-numeric/jquery-numeric.js'], function () {
            numeric.numericInput()
        });
    }

    var date = $(parent).find("input[upgrade-type=date],input[upgrade-type=datetime],input[upgrade-type=daterange]")
    if(date.length > 0) {
        require(
            'plugins/moment/moment.min.js',
            'plugins/daterangepicker/daterangepicker.js',
            'plugins/daterangepicker/daterangepicker.css',
            function () {
                $(date).each(function () {
                    params = UPGRADE_DEFAULT_DATE
                    if($(this).attr("min")){
                        params["minDate"] = $(this).attr("min")
                    }
                    if($(this).attr("max")){
                        params["maxDate"] = $(this).attr("max")
                    }
                    if($(this).attr("min-year")){
                        params["minYear"] = $(this).attr("min-year")
                    }
                    if($(this).attr("max-year")){
                        params["maxYear"] = $(this).attr("max-year")
                    }

                    if($(this).attr("format")){
                        params["locale"]["format"] = $(this).attr("format")
                    }

                    switch($(this).attr("upgrade-type")) {
                        case "date":
                            params["singleDatePicker"] = true
                            break;
                        case "datetime":
                            params["singleDatePicker"] = true
                            params["timePicker"] = true
                            params["locale"]["format"] = ($(this).attr("format"))?$(this).attr("format"):"YYYY/MM/DD hh:mm A";
                            break;
                        case "daterange":
                            params["singleDatePicker"] = false
                            params["timePicker"] = false
                            params["locale"]["format"] = ($(this).attr("format"))?$(this).attr("format"):"YYYY/MM/DD";
                            break
                        case "datetimerange":
                            params["singleDatePicker"] = false
                            params["timePicker"] = true
                            params["locale"]["format"] = ($(this).attr("format"))?$(this).attr("format"):"YYYY/MM/DD hh:mm A";
                            break
                    }

                    $(this).daterangepicker(params);
                })

        });
    }

    var masked = $(parent).find("[masked]")
    if(masked.length > 0){
        require(
            "plugins/inputmask/min/jquery.inputmask.bundle.min.js",
            function () {
                masked.inputmask();
            });
    }

    var switches = $(parent).find("[data-bootstrap-switch]")
    if(switches.length > 0){
        require(
            "plugins/bootstrap-switch/js/bootstrap-switch.min.js",
            function () {
                switches.each(function(){
                    $(this).bootstrapSwitch('state', $(this).prop('checked'));
                });
            });
    }


}