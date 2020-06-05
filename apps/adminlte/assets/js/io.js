var IO = {}

IO.defaultCookieAttributes = { expires: 365, path:"", }

IO.URL = function () {
    return new URL()
}

class URL {
    constructor(url) {
        if (!url) {
            this.raw = window.location.href.toString();
            this.hostname = window.location.hostname
            this.origin = window.location.origin
            this.scheme = window.location.protocol.split(":")[0]
            this.path = window.location.pathname
            if (window.location.search) {
                this.query = this.parseQuery(window.location.search.substring(1))
            } else {
                this.query = {}
            }

        } else {
            this.raw = url
        }

    }

    parseQuery(s) {
        return JSON.parse('{"' + decodeURI(s.replace(/&/g, "\",\"").replace(/=/g, "\":\"")) + '"}')
    }

    set(key, val) {
        this.query[key] = val
        return this
    }

    remove(key, val) {
        if (this.query.hasOwnProperty(key)) {
            delete this.query[key]
        }
        return this
    }

    get(key) {
        if (this.query.hasOwnProperty(key)) {
            return this.query[key]
        }
        return false
    }

    build() {
        var str = [];
        for (var p in this.query)
            if (this.query.hasOwnProperty(p)) {
                str.push(encodeURIComponent(p) + "=" + encodeURIComponent(this.query[p]));
            }
        return this.raw.split("?")[0] + "?" + str.join("&");
    }
}


/**
 * Initialize the IO preparations on each load
 */
IO.Init = function(){
    $.ajaxSetup({
        headers: { 'X-Request-Type': 'application/io' , "Accept":"application/json"}
    });
}()

IO.defaultAjaxAttributes = {dataType:"JSON", data:[], type:"POST", error:function (error) { if(IO.onAjaxError != undefined) IO.onAjaxError(error);}}
IO.Ajax = function(params){
    params = IO.Assign({},IO.defaultAjaxAttributes,params)
    $.ajax(params)
}

/**
 * Get/Set Cookie
 * @param key determine the key, if nothing passes it will return all cookies as json object
 * @param set is value to set, if nothing pass depend on key function return a or all cookies
 * @param attributes is cookie attributes including expire,domain,secure,path
 */
IO.Cookie = function (key,set,attributes) {
  if (typeof document === 'undefined' || (arguments.length && !key)) {
     return
  }

  if(set == undefined){
      // To prevent the for loop in the first place assign an empty array
      // in case there are no cookies at all.
      var cookies = document.cookie ? document.cookie.split('; ') : []
      var jar = {}
      for (var i = 0; i < cookies.length; i++) {
          var parts = cookies[i].split('=')
          var value = parts.slice(1).join('=')
          var foundKey = parts[0].replace(/%3D/g, '=').replace(/%3B/g, ';')
          jar[foundKey] = value.replace(/%3D/g, '=').replace(/%3B/g, ';')

          if (key === foundKey) {
              break
          }
      }

      return key ? jar[key] : jar
  }else{

      console.warn(typeof set)

      attributes = IO.Assign({}, IO.defaultCookieAttributes, attributes)
      if (typeof attributes.expires === 'number') {
          attributes.expires = new Date(Date.now() + attributes.expires * 864e5)
      }
      if (attributes.expires) {
          attributes.expires = attributes.expires.toUTCString()
      }

      key = key.replace(/=/g, '%3D').replace(/;/g, '%3B')

      value = String(set).replace(/=/g, '%3D').replace(/;/g, '%3B')

      var stringifiedAttributes = ''
      for (var attributeName in attributes) {
          if (!attributes[attributeName]) {
              continue
          }

          stringifiedAttributes += '; ' + attributeName

          if (attributes[attributeName] === true) {
              continue
          }

          stringifiedAttributes += '=' + attributes[attributeName].split(';')[0]
      }

      return (document.cookie = key + '=' + value + stringifiedAttributes)
  }
}

/**
 * Remove Cookie
 * @param key determine the key to delete
 */
IO.RemoveCookie = function(key){
    if (typeof document === 'undefined' || (arguments.length && !key)) {
        return
    }
    IO.Cookie(key,"",{expires:-100000000 , path: "/"})
}

/**
 * Show alert dialog on page
 * @param message message to show on dialog box
 * @param onclose optional dialog callback
 */
IO.Alert = function(message,onclose){

}

/**
 * Ask for user confirmation
 * @param message message to show on dialog box
 * @param yes optional onYes click callback
 * @param no optional onNo click callback
 */
IO.Confirm = function(message,yes,no){

}

IO.Loading = function(timeout){

}

/**
 * Set message to show on next loading page
 * @param message is message to show on page
 * @param type is message type (default:info)
 */
IO.Message = function(message,type){
    if(message == undefined){
        return
    }
    if(type == undefined){
        type = "info"
    }
}

/**
 * Redirect to another page with given message
 * @param message optional message to show on page
 * @param type is message type (default:info)
 */
IO.Redirect = function(url,message,type){
    if(message != undefined){
        IO.Message(message,type)
    }
    window.location = IO.Route(url)
}

/**
 * Go back to previous page with given message
 * @param message optional message to show on page
 * @param type is message type (default:info)
 */
IO.Back = function(message,type){
    if(message != undefined){
        IO.Message(message,type)
    }
    window.history.back()
}

/**
 * Refresh page with given message
 * @param message optional message to show on page
 * @param type is message type (default:info)
 */
IO.Refresh = function(message,type){
    if(message != undefined){
        IO.Message(message,type)
    }
    window.location.reload()
}

/**
 * Return absolute uri to relative route
 */
IO.Route = function(path){
    r = new RegExp('^(?:[a-z]+:)?//', 'i');
    if(!r.test(path)){
        path = IO.Base().trim("/")+"/"+path
    }
    return path;
}

/**
 * Return base path
 */
IO.Base = function(){
    uri = window.location;
    parts = uri.pathname.split('/')
    base = uri.protocol + "//" + uri.host + "/" + parts[1];
    if(parts[2] == "admin"){
        base += "/admin";
    }
    return base
}

/**
 * Take input args and assign to default values
 */
IO.Assign = function (target) {
    for (var i = 1; i < arguments.length; i++) {
        var source = arguments[i]
        for (var key in source) {
            target[key] = source[key]
        }
    }
    return target
}

/**
 * Return formatted text , like sprintf
 */
String.prototype.format = function() {
    var args = arguments;
    if(args.length == 1 && Array.isArray(args[0]))
        args = args[0]
    return this.replace(/{(\d+)}/g, function(match, number) {
        return typeof args[number] != 'undefined'
            ? args[number]
            : match
            ;
    });
};



/**
 * Return true if string cintains substring
 * @param str substring to search
 * @param position start position to search
 */
String.prototype.contains = function(str,position) {
    this.includes(str,position)
};


/**
 * Return true if string is numeric and alphabets
 */
String.prototype.isAlphaNumeric = function() {
    var code, i, len;
    for (i = 0, len = this.length; i < len; i++) {
        code = this.charCodeAt(i);
        if (!(code > 47 && code < 58) && // numeric (0-9)
            !(code > 64 && code < 91) && // upper alpha (A-Z)
            !(code > 96 && code < 123)) { // lower alpha (a-z)
            return false;
        }
    }
    return true;
};

/**
 * Return true if string is alphabets
 */
String.prototype.isAlpha = function() {
    var code, i, len;
    for (i = 0, len = this.length; i < len; i++) {
        code = this.charCodeAt(i);
        if (!(code > 64 && code < 91) && // upper alpha (A-Z)
            !(code > 96 && code < 123)) { // lower alpha (a-z)
            return false;
        }
    }
    return true;
};

/**
 * Return true if string is numeric
 */
String.prototype.isNumeric = function() {
    var code, i, len;
    for (i = 0, len = this.length; i < len; i++) {
        code = this.charCodeAt(i);
        if (!(code > 47 && code < 58)) { // lower alpha (a-z)
            return false;
        }
    }
    return true;
};

/**
 * Return int value of string
 */
String.prototype.int = function() {
    return parseInt(this)
};

/**
 * Return float value of string
 */
String.prototype.float = function() {
    return parseFloat(this)
};

/**
 * Return boolean value of string
 * t , true, yes, y, on, 1 considered as true
 */
String.prototype.boolean = function () {
    s = this.toLowerCase()
    return s == "t" || s == "true" || s == "yes" || s == "y" || s == "on" || s == "1"
}

/**
 * Slugify text
 */
String.prototype.slugify = function() {
    str = this.replace(/^\s+|\s+$/g, ''); // trim
    str = str.toLowerCase();

    // remove accents, swap ñ for n, etc
    var from = "ãàáäâẽèéëêìíïîõòóöôùúüûñç·/_,:;";
    var to   = "aaaaaeeeeeiiiiooooouuuunc------";
    for (var i=0, l=from.length ; i<l ; i++) {
        str = str.replace(new RegExp(from.charAt(i), 'g'), to.charAt(i));
    }

    str = str.replace(/[^a-z0-9 -]/g, '') // remove invalid chars
        .replace(/\s+/g, '-') // collapse whitespace and replace by -
        .replace(/-+/g, '-'); // collapse dashes

    return str;
};

/**
 * Replace template tags with data inside string
 * @param data provided json object to replace in text
 */
String.prototype.template = function (data) {
    return this.replace(/\{([\w\.]*)\}/g, function(str, key) {
        var keys = key.split("."), v = data[keys.shift()];
        for (var i = 0, l = keys.length; i < l; i++) v = v[keys[i]];
        return (typeof v !== "undefined" && v !== null) ? v : "";
    });
}

/**
 * Sprintf function
 */
String.prototype.sprintf = function () {
    var args = arguments;
    if(args.length == 1 && Array.isArray(args[0]))
        args = args[0]
    return this.replace(/{(\d+)}/g, function(match, number) {
        return typeof args[number] != 'undefined'
            ? args[number]
            : match
            ;
    });
}

/**
 * Truncate text to certain length
 * @param m is length of text
 * @param ending is ending of text if truncated
 */
String.prototype.truncate = function (m,ending) {
    if(ending == undefined){
        ending = ""
    }
    return (this.length > m)
        ? this.trim().substring(0, m).split(" ").slice(0, -1).join(" ") + ending
        : this;
}

/**
 * Convert text to title case
 */
String.prototype.toTitleCase = function () {
    return this.replace(/(?:^|\s)\w/g, function(match) {
        return match.toUpperCase();
    });
}

String.prototype.trimChars = function (c) {
    var re = new RegExp("^[" + c + "]+|[" + c + "]+$", "g");
    return this.replace(re,"");
}

/**
 * Convert text to under score case
 */
String.prototype.toUnderScoreCase = function () {
    return this.replace(/\.?([A-Z])/g, function (x,y){return "_" + y.toLowerCase()}).replace(/^_/, "")
}

/**
 * Convert text to camel case
 */
String.prototype.toCamelCase = function () {
    return this.replace(/(?:^\w|[A-Z]|\b\w)/g, function(word, index) {
        return index === 0 ? word.toLowerCase() : word.toUpperCase();
    }).replace(/\s+/g, '');
}

/**
 * Add separator every 3 digit
 * @param sep A value used to specify separator character, by default its comma
 */
String.prototype.thousands = function (sep) {
    if(sep == undefined){
        sep = ","
    }
    return this.toString().replace(/\B(?=(\d{3})+(?!\d))/g, sep);
}

/**
 * Strip any html tags from text
 */
String.prototype.stripTags = function () {
    var div = document.createElement("div");
    div.innerHTML = this;
    return  div.textContent || div.innerText || "";
}


/**
 * Split a string into lines using the newline separator and return them as an array.
 * @param limit A value used to limit the number of elements returned in the array.
 */
String.prototype.lines = function (limit){
    return this.split("\n",limit)
};

/**
 * Computes a new string in which certain characters have been replaced by a hexadecimal escape sequence.
 * @param string A string value
 */
String.prototype.escape = function (){
    return escape(this)
};

/**
 * Computes a new string in which hexadecimal escape sequences are replaced with the character that it represents.
 * @param string A string value
 */
String.prototype.unescape = function (){
    return unescape(this)
};

String.prototype.json = function (){
    return JSON.parse(this)
};

/**
 * Remove item from array by value
 * @param value of member to remove
 */
Array.prototype.remove = function(member) {
    var index = this.indexOf(member);
    if (index > -1) {
        this.splice(index, 1);
    }
    return this;
}

/**
 * Foreach over array
 * @param callbackFn callback function take index,value
 */
Array.prototype.each = function(callbackFn) {
    this.forEach(callbackFn)
}


Array.prototype.contains = function(value) {
    for(i=0; i < this.length; i++){
        if(this[i] == value){
            return true
        }
    }
    return false
}

Array.prototype.last = function() {
    return this[this.length - 1]
}

Array.prototype.first = function() {
    return this[0]
}

Array.prototype.rest = function(values) {
    res = []
    for(i=0; i < this.length; i++){
        found = false
        for(j=0; j < values.length; j++){
            if(this[i] == value){
                found = true
                break
            }
        }
        if(!found){
            res.push(this[i])
        }
    }
    return res
}


Array.prototype.prepend = function(v){
    return this.unshift(v)
}

Array.prototype.append = function(v){
    return this.push(v)
}

Array.prototype.pluck = function(key){
    res = []
    for(i=0; i < this.length; i++)
    if(this[i].hasOwnProperty(key)){
        res.push(this[key])
    }
    return res
}

Array.prototype.min = function(){
    return Math.min.apply(null,this)
}

Array.prototype.max = function(){
    return Math.max.apply(null,this)
}

Array.prototype.avg = function(){
    res = 0
    for(i=0; i < this.length; i++)
       res += this[i]
    res /= this.length
    return res
}

/*Object.prototype.get = function(key){
    if(this.hasOwnProperty(key)){
        return this[key]
    }
    return undefined
}

Object.prototype.set = function(key,value){
    return this[key] = value
}*/

/*
Object.prototype.has = function(key){
    return this.hasOwnProperty(key)
}
*/




/*Object.prototype.string = function(){
    return JSON.stringify(this)
}*/

var _require_loaded = {}
const require = function (...args) {
    var callback = false;
    var url = [];
    var async = false;
    var done = 0;
    for(var i = 0; i < args.length; i++){
        if(Array.isArray(args[i])){
            for(j = 0; j < args[i].length;j++){
                if(typeof args[i][j] === "string" && !_require_loaded.hasOwnProperty(args[i][j])){
                    url.push(args[i][j]);
                }
            }
            continue
        }
        if(typeof args[i] === "string" && !_require_loaded.hasOwnProperty(args[i]) ){
                url.push(args[i]);
        }
        if(typeof args[i] === "function"){
            callback = args[i];
            async = true;
        }
    }
    if(url.length == 0 && async == true){
        callback();
    }
    for(var i = 0; i < url.length; i++){

        parts = url[i].split("/")
        file = parts[parts.length - 1]
        parts = file.split(".")
        ext = parts[parts.length - 1]
        if(ext == "css") {
            var el = jQuery("<link>", {rel: "stylesheet", type: "text/css", href: url[i]})[0];
            el.onload = function () {
                if (async) {
                    done++
                    if (done == url.length) {
                        callback()
                    }
                }
            }
            document.getElementsByTagName("head")[0].appendChild(el);
        }

        if(ext == "js") {
            jQuery.ajax({
                url: url[i],
                method: "GET",
                dataType: "script",
                async: async,
                cache: true,
                success: function () {
                    if (async) {
                        done++
                        if (done == url.length) {
                            callback()
                        }
                    }
                },
                error: function () {
                    console.error("could not load %s", url[i])
                    if (async) {
                        done++
                        if (done == url.length) {
                            callback()
                        }
                    }
                }
            });
        }

    }

};