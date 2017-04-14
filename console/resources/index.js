var appId, appKey, masterKey, hookKey, leanenginePort, sendHookKey ;
var hooksInfo = {};
var functionsInfo = [];
var warnings = [];
var _ = AV._;

$(document).ready(function (){
  $.fn.select2.defaults.set("theme", "bootstrap");
  $('#functions').select2();
  $('#hooks').select2();
  $('#classes').select2();

  $.get("/__engine/1/appInfo", function(data) {
    appId = data.appId;
    appKey = data.appKey;
    masterKey = data.masterKey;
    hookKey = data.hookKey;
    sendHookKey = data.sendHookKey;
    leanenginePort = data.leanenginePort;
    AV._initialize(appId, appKey, masterKey);
    AV._useMasterKey = true;

    for (var i=0; i<data.warnings.length; i++) {
      warnings.push(data.warnings[i]);
    }
  });

  $(window).onerror = function (msg, url, line){
    $('#result').html(msg);
  };

  //load options
  $("#classes").change(function (event){
    var className = $(this).val();
    if(className != null && className != ''){
      $('#hooks').find('option').remove().end();
      $.get("/__engine/1/classes/" + className + "/actions", function (data){
        hooksInfo[className] = data;
        _.each(data, function (name){
          $('#hooks').append($("<option></option>")
                             .attr("value", "/" + name.className + "/" + name.action)
                             .text(name.action));
        });
        $("#hooks").trigger('change');
      });
    }
  });

  $('#hooks').change(function() {
    if ($('#hooks').val().match(/(before|after)Update$/))
      $('#divUpdatedKeys').show();
    else
      $('#divUpdatedKeys').hide();
  });

  $.get("/__engine/1/functions", function (data){
    functionsInfo = data;
    _.each(data, function (info) {
      $('#functions')
        .append($("<option></option>")
                .attr("value", info.name)
                .text(info.name));
    });
  });

  $.get("/__engine/1/classes", function (data){
    _.each(data, function (name){
      $('#classes')
        .append($("<option></option>")
                .attr("value", name)
                .text(name));
    });
    $("#classes").trigger('change');
  });

  var request = function(url, data, user) {
    var data = data || {}
    $.ajax({
      type       : "POST",
      url        : url,
      headers    : {
        "X-AVOSCloud-Application-Id": appId,
        "X-AVOSCloud-Application-Key": appKey,
        "X-AVOSCloud-Session-Token": user ? user._sessionToken : undefined,
        "X-LC-Hook-Key": sendHookKey ? hookKey : undefined
      },
      data       : JSON.stringify(data),
      dataType   : 'json',
      contentType: 'application/json',
      success    : function (data){
        $('#result').html(JSON.stringify(data, null, '  '));
      },
      error      : function(xhr) {
        $('#result').html(xhr.responseText);
      }
    });
  }

  var getUser = function(uid, cb) {
    if (uid && uid.trim() != '') {
      var user = AV.Object.createWithoutData("_User", uid);
      user.fetch({
        success: function(user) {
          cb(null, user);
        },
        error: function(user, err) {
          cb(err);
        }
      });
    } else {
      return cb(null, null);
    }
  }

  function parseJSON(s){
    return JSON.stringify(eval('(' + s + ')'));
  }

  //events
  $('#callFuncs').click(function (e){
    try {
      $('#result').html('');
      var paramsStr = $('#params').val();
      var data = null;
      if(paramsStr != null && paramsStr.trim() != ''){
        data = JSON.parse(parseJSON(paramsStr));
      }

      if (!sendHookKey && _.contains([ // Node SDK < 2.0
        '_messageReceived', '_receiversOffline', '_messageSent', '_conversationStart', '_conversationStarted',
        '_conversationAdd', '_conversationRemove', '_conversationUpdate'
      ], $('#functions').val())) {
        data = data || {};
        data.__sign = _.findWhere(functionsInfo, {name: $('#functions').val()}).sign;
      }

      getUser($('#userId1').val(), function(err, user) {
        if (err) {
          return $('#result').html(err.message || err);
        }
        var apiEndpoint = $('#isCall').is(':checked') ? '/1.1/call/' : '/1.1/functions/';
        var url = "http://" + window.location.hostname + ":" + leanenginePort + apiEndpoint + $('#functions').val();
        request(url, data, user);
      })
    } catch(e){
      $('#result').html(e.message);
      return null;
    }
  });

  $('#callHooks').click(function (e){
    try {
      $('#result').html('');
      var getObject = function(cb) {
        var objStr = $('#object').val().trim();
        var objectId = $('#objectId').val().trim();
        if(objStr && objStr != ''){
          return cb(null, { object: JSON.parse(parseJSON(objStr)) });
        } else if (objectId && objectId != ''){
          var className = $('#classes').val();
          object = AV.Object.createWithoutData(className, objectId);
          object.fetch().then(function(obj) {
            if (obj.createdAt) {
              cb(null, { object: JSON.parse(JSON.stringify(obj)) });
            } else {
              cb(new Error('Could not find ' + className + ' object, objectId: ' + objectId))
            }
          }, function(err) {
            cb(err)
          });
        } else {
          cb(null, { object: {} });
        }
      }
      getUser($('#userId2').val(), function(err, user) {
        if (err) {
          return $('#result').html(err.message || err);
        }
        var url = window.location.protocol + '//' + window.location.hostname + ':' + leanenginePort + "/1.1/functions" + $('#hooks').val();
        getObject(function(err, data) {
          var hookName = $('#hooks :selected').text();

          if (err) {
            return $('#result').html(err.message || err);
          }

          if ($('#hooks').val().match(/(before|after)Update$/) && $('#updatedKeys').val()) {
            data.object._updatedKeys = $('#updatedKeys').val().split(/,\s+/);
          }

          var hookInfo = _.findWhere(hooksInfo[$('#classes').val()], {action: hookName});

          if (!sendHookKey) { // Node SDK < 2.0
            var sign = hookInfo && hookInfo.sign;

            if (hookName.indexOf('before') === 0) {
              data.object.__before = sign;
            } else if (hookName.indexOf('after') === 0) {
              data.object.__after = sign;
            } else {
              data.object.__sign = sign;
            }
          }

          if (user) {
            data.user = user.toJSON();
            data.user.sessionToken = user._sessionToken;
          }

          request(url, data, user);
        });
      })
    } catch(e){
      $('#result').html(e.message);
    }
  });

  new Vue({
    el: "#warnings",
    data: {
      warnings: warnings,
    },
  });
});
