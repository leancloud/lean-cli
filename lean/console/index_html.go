package console

const indexHTML = `
<!DOCTYPE html>
<html>
<head>
  <script type="text/javascript" src="http://cdn.staticfile.org/jquery/2.0.3/jquery.min.js"></script>
  <script type="text/javascript" src="http://cdn.staticfile.org/underscore.js/1.5.2/underscore-min.js"></script>
  <script type="text/javascript" src="https://cdn1.lncld.net/static/js/av-core-mini-0.5.4.js"></script>
  <link rel="stylesheet" href="http://cdn.staticfile.org/twitter-bootstrap/3.3.1/css/bootstrap.min.css" type="text/css" media="screen"/>
  <script type="text/javascript" charset="utf-8">
    var appId, appKey, masterKey, leanenginePort;
    var hooksInfo = {};
    var functionsInfo = [];

    $(document).ready(function (){

      $.get("/__engine/1/appInfo", function(data) {
        appId = data.appId;
        appKey = data.appKey;
        masterKey = data.masterKey;
        leanenginePort = data.leanenginePort;
        AV._initialize(appId, appKey, masterKey);
        AV._useMasterKey = true;
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
            "X-AVOSCloud-Session-Token": user ? user._sessionToken : undefined
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
          if (_.contains([
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

              var sign = hookInfo && hookInfo.sign;

              if (hookName.indexOf('before') === 0) {
                data.object.__before = sign;
              } else if (hookName.indexOf('after') === 0) {
                data.object.__after = sign;
              } else {
                data.object.__sign = sign;
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
    });
  </script>
  <title>LeanCloud Cloud Code Debug Tool</title>
</head>
<body>
<div class="container-fluid">
  <div class="row-fluid">
    <div class="col-md-4 col-md-offset-2">
      <h4>测试 AV.Cloud.define 的函数</h4>
      <div class="form-group">
        <label>选择函数</label>
        <select id="functions" class="form-control"></select>
      </div>
      <div class="form-group">
        <label>登录用户的 Object Id（模拟登录用户，可为空）</label>
        <input id="userId1" name="userId1" class="form-control"></input>
      </div>
      <div class="form-group">
        <label>传入 JSON 格式参数（可为空）</label>
        <span>
          &raquo; <input type="checkbox" id="isCall">
          <label for="isCall">作为 AVObject 传输（call）</label>
        </span>
        <textarea id='params' rows="10" cols="80" class="form-control"></textarea>
      </div>
      <button type="button" id="callFuncs" class="btn btn-default">执行</button>
    </div>

    <div class="col-md-4">
      <h4>测试 Class Hooks（beforeSave, afterSave 等）</h4>
      <div class="form-group">
        <label>选择 Class</label>
        <select id="classes" class="form-control"></select>
      </div>
      <div class="form-group">
        <label>选择函数</label>
        <select id="hooks" class="form-control"></select>
      </div>
      <div class="form-group">
        <label>登录用户的 Object Id（模拟登录用户，可为空）</label>
        <input id="userId2" name="userId2" class="form-control"></input>
      </div>
      <div class="form-group">
        <label>填写已经存在对象的 objectId</label>
        <input type="text" id="objectId" class="form-control"></input>
      </div>
      <div class="form-group" id="divUpdatedKeys">
        <label>修改过的字段（逗号隔开）</label>
        <input type="text" id="updatedKeys" class="form-control"></input>
      </div>
      <div class="form-group">
        <label>或者传入 JSON 格式的对象</label>
        <textarea rows="10" id="object" class="form-control"></textarea>
      </div>
      <button type="button" id="callHooks" class="btn btn-default">执行</button>
    </div>
  </div>
  <div class="row-fluid">
    <div class='col-md-8 col-md-offset-2 alert alert-info'>
      <h4>结果</h4>
      <pre id='result'></pre>
    </div>
  </div>
</div>
<div>
  <footer class="footer">
  <div class="row-fluid text-center">
    <p>&copy;<a href="https://leancloud.cn">LeanCloud</a>, All Rights Reserved.</p>
  </div>
  </footer>
</div>
</body>
</html>
`
