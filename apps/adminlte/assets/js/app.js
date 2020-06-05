var App = {};


App.WindowId = 0
App.Alert = function (message,callback) {
    alert(message)
    if(callback){
        callback()
    }
}



IO.onAjaxError = function (error) {
    IO.Loading(false);
    App.Alert("Unable to complete the request.")
};

App.BreadCrumb = function () {
    var breadcrumb = [];
    try{
        var  json = JSON.parse(window.name)
        App.WindowId = json.id
        breadcrumb = json.breadcrumb;
    }catch (e) {
        App.WindowId = new Date().getTime()
    }

    var title = $("[page-title]").text().trim()

    if(title != "") {

        if( breadcrumb.length == 0 || (breadcrumb.length > 0 && breadcrumb[ breadcrumb.length - 1].title != title)) {
            breadcrumb.push({
                "url": window.location.href,
                "title": title
            })
            if (breadcrumb.length > 4) {
                breadcrumb.shift()
            }
        }
        window.name = JSON.stringify({id:App.WindowId,breadcrumb:breadcrumb});
    }
    var el = $("[page-breadcrumb]")
    if(el.length != 0){
        for(var i = 0; i < breadcrumb.length - 1; i++){
            el.append("<li class=\"breadcrumb-item\"><a href='"+breadcrumb[i].url+"'>"+breadcrumb[i].title+"</a></li>");
        }
        el.append("<li class=\"breadcrumb-item active\">"+breadcrumb[breadcrumb.length - 1].title+"</li>");
    }
}();

App.Upgrade = function () {
    jQuery("[upgrade]").each(function () {
       el = jQuery(this);
       if(el.is("form")){
           Upgrade(el);
       }
   }) 
}();

App.Init = function () {
    jQuery('a[data-toggle="tab"],a[data-toggle="pill"]').on('shown.bs.tab', function(e) {
        localStorage.setItem('activeTab-'+$(this).closest("[role=tablist]").attr("id"), $(e.target).attr('href'));
    });
    jQuery("[role=tablist]").each(function () {

        var activeTab = localStorage.getItem('activeTab-'+$(this).attr("id"));
        if(activeTab){
            jQuery('#'+$(this).attr("id")+' a[href="' + activeTab + '"]').tab('show');
        }
    });

    $("[onpressenter]").on('keypress', function (e) {
        if (e.which == 13) {
            eval($(this).attr("onpressenter"))
        }
    });

}();

App.Error = function (res) {
    if(res.success){
        return false
    }
    console.error(res)
    return true
};


App.Flash = function () {
    IO.RemoveCookie("flash")
}()