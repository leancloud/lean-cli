/* globals window, document, AV, $, Vue */
"use strict";

var _ = AV._;

// init select2
$.fn.select2.defaults.set("theme", "bootstrap");

// create select-vue component
Vue.component('select2', {
  props: ['options', 'value'],
  template: '#select2-template',
  mounted: function () {
    var vm = this;
    $(this.$el)
      // init select2
      .select2({ data: this.options })
      .val(this.value)
      .trigger('change')
      // emit event on change.
      .on('change', function () {
        vm.$emit('change');
        vm.$emit('input', this.value);
      });
  },
  watch: {
    value: function (value) {
      // update value
      $(this.$el).val(value).trigger('change');
    },
    options: function (options) {
      // update options
      $(this.$el).select2('destroy').empty();
      $(this.$el).select2({ data: options });
    }
  },
  destroyed: function () {
    $(this.$el).off().select2('destroy');
  }
});

function isRtmFunction(funcName) {
  return _.contains([
    '_messageReceived', '_receiversOffline', '_messageSent', '_conversationStart', '_conversationStarted',
    '_conversationAdd', '_conversationRemove', '_conversationUpdate'
  ], funcName);
}

function getAppInfo() {
  return $.getJSON("/__engine/1/appInfo").then(function(info) {
    AV.init({ appId: info.appId, appKey: info.appKey, masterKey: info.masterKey});
    // TODO: don't use private function
    AV._config.useMasterKey = true;
    return info;
  });
}

function getCloudFunction() {
  return $.get("/__engine/1/functions").then(function(cloudFunctions) {
    _.each(cloudFunctions, function(cloudFunction, idx) {
      cloudFunction.id = idx;
      cloudFunction.text = cloudFunction.name;
    });
    return cloudFunctions;
  });
}

function getHookFunctions(className) {
  return $.get("/__engine/1/classes/" + className + "/actions").then(function(hookFunctions) {
    _.each(hookFunctions, function(hookFunction, idx) {
      hookFunction.id = idx;
      hookFunction.text = hookFunction.action;
    });
    return hookFunctions;
  });
}

function getHookClasses() {
  return $.get("/__engine/1/classes").then(function(hookClasses) {
    return _.map(hookClasses, function(hookClass, idx) {
      return {
        id: idx,
        name: hookClass,
        text: hookClass,
      };
    });
  });
}

function getUserHooks() {
  return $.get("/__engine/1/userHooks").then(function(hooks) {
    return _.indexBy(hooks, 'action')
  });
}

function getUser(uid) {
  if (!uid || uid.trim() === '') {
    return AV.Promise.resolve(null);
  }
  var user = AV.Object.createWithoutData("_User", uid);
  return user.fetch();
}

function callCloudFunction(appInfo, cloudFunction, params, user, isCall) {
  var data = null;
  if(params !== null && params.trim() !== '') {
    try {
      data = JSON5.parse(params);
    } catch (err) {
      var dtd = $.Deferred();
      dtd.reject(err);
      return dtd;
    }
  }

  var apiEndpoint = isCall ? '/1.1/call/' : '/1.1/functions/';
  var url = appInfo.remoteUrl + apiEndpoint + cloudFunction.name;
  data = data || {};

  if (!appInfo.sendHookKey && isRtmFunction(cloudFunction.name)) {
    data.__sign = cloudFunction.sign;
  }

  return $.ajax({
    type: "POST",
    url: url,
    headers: {
      "X-AVOSCloud-Application-Id": appInfo.appId,
      "X-AVOSCloud-Application-Key": appInfo.appKey,
      "X-AVOSCloud-Session-Token": user ? user._sessionToken : undefined,
      "X-LC-Hook-Key": appInfo.sendHookKey ? appInfo.hookKey : undefined
    },
    data: JSON.stringify(data),
    dataType: 'json',
    contentType: 'application/json',
  });
}

function getHookObjectById(className, objId) {
  return new AV.Query(className).get(objId).then(function(obj) {
    // make this AV.Object to a plain Object
    return JSON.parse(JSON.stringify(obj));
  });
}

function getHookObjectByContent(content) {
  return new AV.Promise(function(resolve, reject) {
    try {
      resolve(JSON5.parse(content));
    } catch (err) {
      reject(err);
    }
  });
}

function callCloudHook(appInfo, hookInfo, data, user) {
  var url = appInfo.remoteUrl + "/1.1/functions/" + hookInfo.className + "/" + hookInfo.action;
  if (user) {
    data.user = user.toJSON();
    data.user.sessionToken = user._sessionToken;
  }

  if (!appInfo.sendHookKey) {
    if (hookInfo.action.match(/^before/)) {
      data.object.__before = hookInfo.sign;
    } else if (hookInfo.action.match(/^after/)) {
      data.object.__after = hookInfo.sign;
    } else {
      data.__sign = hookInfo.sign;
    }
  }

  return $.ajax({
    type: "POST",
    url: url,
    headers: {
      "X-AVOSCloud-Application-Id": appInfo.appId,
      "X-AVOSCloud-Application-Key": appInfo.appKey,
      "X-AVOSCloud-Session-Token": user ? user._sessionToken : undefined,
      "X-LC-Hook-Key": appInfo.sendHookKey ? appInfo.hookKey : undefined
    },
    data: JSON.stringify(data),
    dataType: 'json',
    contentType: 'application/json',
  });
}

function addToHistoryOperations(operations, operation) {
  for (var i=0; i<operations.length; i++) {
    if (_.isEqual(operations[i], operation)) {
      // operation already exists, just remove the old one and add new one to top.
      operations.splice(i, 1);
      operations.unshift(operation);
      window.localStorage.leanCliOpHistory = JSON.stringify(operations);
      return;
    }
  }

  operations.unshift(operation);
  if (operations.length > 30) {
    operations.pop();
  }
  window.localStorage.leanCliOpHistory = JSON.stringify(operations);
}

$(document).ready(function() {
  new Vue({
    el: "#application",
    data: {
      appInfo: {},
      warnings: [],
      result: '',

      selectedPanel: 0,

      // cloud function related:
      cloudFunctions: [],
      selectedFunction: 0,
      cloudFunctionUserId: null,
      isCall: false,
      cloudFunctionParams: null,

      // cloud hook related:
      hookClasses: [],
      hookFunctions: [],
      hookObjectId: null,
      hookObjectContent: null,
      hookUserId: null,
      updatedKeys: '',
      selectedClass: 0,
      selectedHook: 0,

      userHooks: [],
      authData: null,
      onVerifiedUserId: null,
      onVerifiedType: null,
      onLoginUserId: null,

      // history related:
      showHistoryPanel: false,
      historyOperations: window.localStorage.leanCliOpHistory ? JSON.parse(window.localStorage.leanCliOpHistory) : [],
    },
    methods: {
      executeCloudFunction: function() {
        getUser(this.cloudFunctionUserId).then((function(user) {
          var cloudFunction = this.cloudFunctions[this.selectedFunction];
          addToHistoryOperations(this.historyOperations, {
            type: 'cloudFunction',
            name: cloudFunction.text,
            userId: this.cloudFunctionUserId,
            params: this.cloudFunctionParams,
          });
          return callCloudFunction(
            this.appInfo,
            cloudFunction,
            this.cloudFunctionParams,
            user,
            this.isCall);
        }).bind(this)).then((function(result) {
          this.result = result;
        }).bind(this)).catch((function(err) {
          this.result = err.responseText || err.message;
        }).bind(this));
      },
      refreshHookFunctions: function() {
        var className = this.hookClasses[this.selectedClass].name;
        getHookFunctions(className).then((function(hookFunctions) {
          this.hookFunctions = hookFunctions;
          this.selectedHook = 0;
        }).bind(this));
      },
      executeCloudHook: function() {
        var hookInfo = this.hookFunctions[this.selectedHook];
        var getObject;
        if (this.hookObjectId !== null && this.hookObjectId.trim() !== "") {
          getObject = (function() { return getHookObjectById(hookInfo.className, this.hookObjectId.trim()); }).bind(this);
        } else if (this.hookObjectContent !== null && this.hookObjectContent.trim() !== "") {
          getObject = (function() { return getHookObjectByContent(this.hookObjectContent); }).bind(this);
        } else {
          getObject = function() { return AV.Promise.resolve({}); };
        }
        addToHistoryOperations(this.historyOperations, {
          type: 'cloudHook',
          className: hookInfo.className,
          hookName: hookInfo.action,
          userId: this.hookUserId,
          objId: this.hookObjectId,
          objContent: this.hookObjectContent,
        });
        AV.Promise.all([getObject(), getUser(this.hookUserId)]).then((function(results) {
          var obj = results[0];
          var user = results[1];
          if (hookInfo.action.match(/^(before|after)Update$/)) {
            if (this.updatedKeys) {
              var keys = this.updatedKeys.split(/,\s*/);
              keys = _.map(keys, function(key) {
                return key.trim();
              });
              keys = _.filter(keys, function(key){
                return key !== '';
              });
              keys = _.uniq(keys);
              obj._updatedKeys = keys;
            }
          }
          return callCloudHook(this.appInfo, hookInfo, {object: obj}, user);
        }).bind(this)).then((function(result) {
          this.result = result;
        }).bind(this)).catch((function(err) {
          this.result = err.responseText || err.message;
        }).bind(this));
      },
      executeOnAuthData: function() {
        const hookInfo = this.userHooks.onAuthData;
        addToHistoryOperations(this.historyOperations, {
          type: 'onAuthData',
          className: hookInfo.className,
          hookName: hookInfo.action,
          params: this.authData,
        });
        callCloudHook(this.appInfo, hookInfo, {authData: JSON.parse(this.authData)}).then((function(result) {
          this.result = result;
        }).bind(this)).fail((function(err) {
          this.result = err.responseText || err.message;
        }).bind(this));
      },
      executeOnLogin: function() {
        const hookInfo = this.userHooks.onLogin;
        addToHistoryOperations(this.historyOperations, {
          type: 'onLogin',
          className: hookInfo.className,
          hookName: hookInfo.action,
          userId: this.onLoginUserId,
        });
        const that = this;
        getUser(this.onLoginUserId).then(function(user) {
          return callCloudHook(that.appInfo, hookInfo, null, user)
        }).then(function(result) {
          that.result = result;
        }).catch(function(err) {
          that.result = err.responseText || err.message;
        })
      },
      executeOnVerified: function(type) {
        const hookInfo = this.userHooks['onVerified' + type];
        hookInfo.name = 'onVerified/' + type.toLowerCase()
        addToHistoryOperations(this.historyOperations, {
          type: hookInfo.action,
          className: hookInfo.className,
          hookName: hookInfo.action,
          userId: this.onLoginUserId,
        });
        const that = this;
        getUser(this.onLoginUserId).then(function(user) {
          return callCloudFunction(that.appInfo, hookInfo, null, user);
        }).then(function(result) {
          that.result = result;
        }).catch(function(err) {
          that.result = err.responseText || err.message;
        })
      },
      restoryHistory: function(operation) {
        if (operation.type === 'cloudFunction') {
          this.cloudFunctionParams = operation.params;
          this.cloudFunctionUserId = operation.userId;
          for (var i=0; i<this.cloudFunctions.length; i++) {
            if (this.cloudFunctions[i].name === operation.name) {
              this.selectedFunction = i;
            }
          }
        } else if (operation.type === 'cloudHook') {
          this.hookUserId = operation.userId;
          this.hookObjectId = operation.objId;
          this.hookObjectContent = operation.objContent;
          for (var i=0; i<this.hookClasses.length; i++) {
            if (this.hookClasses[i].name === operation.className) {
              this.selectedClass = i;
            }
          }
          for (var i=0; i<this.hookFunctions.length; i++) {
            if (this.hookFunctions[i].action === operation.hookName) {
              this.selectedHook = i;
            }
          }
        } else {
          throw new TypeError('invalid operation type: ' + operation.type);
        }
      },
    },
    mounted: function() {
      $.when(getAppInfo(), getCloudFunction(), getHookClasses(), getUserHooks()).then((function(appInfo, cloudFunctions, hookClasses, userHooks) {
        this.warnings = appInfo.warnings;
        this.appInfo = appInfo;
        this.cloudFunctions = cloudFunctions;
        this.hookClasses = hookClasses;
        this.userHooks = userHooks;

        if (this.hookClasses.length === 0) {
          return [];
        }
        var className = this.hookClasses[this.selectedClass].name;
        return getHookFunctions(className);
      }).bind(this)).then((function(hookFunctions) {
        this.hookFunctions = hookFunctions;
        this.selectedHook = 0;
      }).bind(this));
    },
  });
});
