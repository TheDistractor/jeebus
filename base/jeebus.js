(function() {
  var ng,
    __slice = [].slice;

  ng = angular.module('myApp');

  ng.factory('jeebus', function($rootScope, $q) {
    var attach, connect, detach, processModelUpdate, processRpcReply, rpc, rpcPromises, send, seqNum, store, trackedModels, ws;
    ws = null;
    seqNum = 0;
    rpcPromises = {};
    trackedModels = {};
    processModelUpdate = function(key, value) {
      var info, k, suffix;
      for (k in trackedModels) {
        info = trackedModels[k];
        if (k === key.slice(0, k.length)) {
          suffix = key.slice(k.length);
          if (value) {
            info.model[suffix] = value;
          } else {
            delete info.model[suffix];
          }
        }
      }
      if (!suffix) {
        return console.error("spurious model update", key, value);
      }
    };
    processRpcReply = function(n, result, err) {
      var d, tid, _ref;
      if (rpcPromises[n]) {
        _ref = rpcPromises[n], tid = _ref[0], d = _ref[1];
        clearTimeout(tid);
        if (err) {
          console.error(err);
          return d.reject(err);
        } else {
          return d.resolve(result);
        }
      } else {
        return console.error("spurious rpc reply", n, result, err);
      }
    };
    connect = function(appTag, port) {
      var reconnect;
      if (port == null) {
        port = location.port;
      }
      reconnect = function(firstCall) {
        ws = new WebSocket("ws://" + location.hostname + ":" + port + "/ws", [appTag]);
        ws.onopen = function() {
          return console.log('WS Open');
        };
        ws.onmessage = function(m) {
          if (m.data instanceof ArrayBuffer) {
            console.log('binary msg', m);
          }
          return $rootScope.$apply(function() {
            var data, e, k, v, _i, _len, _ref, _results, _results1;
            data = JSON.parse(m.data);
            if (m.data[0] === '[') {
              return processRpcReply.apply(null, data);
            } else {
              switch (data) {
                case true:
                  return window.location.reload(true);
                case false:
                  console.log("CSS Reload");
                  _ref = document.getElementsByTagName('link');
                  _results = [];
                  for (_i = 0, _len = _ref.length; _i < _len; _i++) {
                    e = _ref[_i];
                    if (e.href && /stylesheet/i.test(e.rel)) {
                      _results.push(e.href = "" + (e.href.replace(/\?.*/, '')) + "?" + (Date.now()));
                    } else {
                      _results.push(void 0);
                    }
                  }
                  return _results;
                  break;
                default:
                  _results1 = [];
                  for (k in data) {
                    v = data[k];
                    _results1.push(processModelUpdate(k, v));
                  }
                  return _results1;
              }
            }
          });
        };
        return ws.onclose = function() {
          console.log('WS Closed');
          return setTimeout(reconnect, 1000);
        };
      };
      return reconnect(true);
    };
    send = function(payload) {
      var msg;
      msg = angular.toJson(payload);
      if (msg[0] === '[') {
        console.error("payload can't be an array (" + payload.length + " elements)");
      } else {
        ws.send(msg);
      }
      return this;
    };
    store = function(key, value) {
      var msg;
      msg = angular.toJson([key, value]);
      if (msg.slice(0, 3) === '["/') {
        ws.send(msg);
      } else {
        console.error('key does not start with "/":', key);
      }
      return this;
    };
    rpc = function() {
      var args, d, n, tid;
      args = 1 <= arguments.length ? __slice.call(arguments, 0) : [];
      d = $q.defer();
      n = ++seqNum;
      ws.send(angular.toJson([n].concat(__slice.call(args))));
      tid = setTimeout(function() {
        console.error("RPC " + n + ": no reponse", args);
        delete rpcPromises[n];
        return $rootScope.$apply(function() {
          return d.reject();
        });
      }, 10000);
      rpcPromises[n] = [tid, d];
      return d.promise;
    };
    attach = function(path) {
      var info;
      info = trackedModels[path] != null ? trackedModels[path] : trackedModels[path] = {
        model: {},
        count: 0
      };
      if (info.count++ === 0) {
        rpc('attach', path).then(function(r) {
          var k, v;
          for (k in r) {
            v = r[k];
            processModelUpdate(k, v);
          }
          return console.log('attach', path);
        });
      }
      return info.model;
    };
    detach = function(path) {
      if (trackedModels[path] && --trackedModels[path].count <= 0) {
        delete trackedModels[path];
        return rpc('detach', path).then(function() {
          return console.log('detach', path);
        });
      }
    };
    return {
      connect: connect,
      send: send,
      store: store,
      rpc: rpc,
      attach: attach,
      detach: detach
    };
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiamVlYnVzLmNvZmZlZSJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUFBLE1BQUEsRUFBQTtJQUFBLGtCQUFBOztBQUFBLEVBQUEsRUFBQSxHQUFLLE9BQU8sQ0FBQyxNQUFSLENBQWUsT0FBZixDQUFMLENBQUE7O0FBQUEsRUFJQSxFQUFFLENBQUMsT0FBSCxDQUFXLFFBQVgsRUFBcUIsU0FBQyxVQUFELEVBQWEsRUFBYixHQUFBO0FBQ25CLFFBQUEsc0hBQUE7QUFBQSxJQUFBLEVBQUEsR0FBSyxJQUFMLENBQUE7QUFBQSxJQUNBLE1BQUEsR0FBUyxDQURULENBQUE7QUFBQSxJQUVBLFdBQUEsR0FBYyxFQUZkLENBQUE7QUFBQSxJQUdBLGFBQUEsR0FBZ0IsRUFIaEIsQ0FBQTtBQUFBLElBTUEsa0JBQUEsR0FBcUIsU0FBQyxHQUFELEVBQU0sS0FBTixHQUFBO0FBQ25CLFVBQUEsZUFBQTtBQUFBLFdBQUEsa0JBQUE7Z0NBQUE7QUFDRSxRQUFBLElBQUcsQ0FBQSxLQUFLLEdBQUcsQ0FBQyxLQUFKLENBQVUsQ0FBVixFQUFhLENBQUMsQ0FBQyxNQUFmLENBQVI7QUFDRSxVQUFBLE1BQUEsR0FBUyxHQUFHLENBQUMsS0FBSixDQUFVLENBQUMsQ0FBQyxNQUFaLENBQVQsQ0FBQTtBQUNBLFVBQUEsSUFBRyxLQUFIO0FBQ0UsWUFBQSxJQUFJLENBQUMsS0FBTSxDQUFBLE1BQUEsQ0FBWCxHQUFxQixLQUFyQixDQURGO1dBQUEsTUFBQTtBQUdFLFlBQUEsTUFBQSxDQUFBLElBQVcsQ0FBQyxLQUFNLENBQUEsTUFBQSxDQUFsQixDQUhGO1dBRkY7U0FERjtBQUFBLE9BQUE7QUFPQSxNQUFBLElBQUEsQ0FBQSxNQUFBO2VBQUEsT0FBTyxDQUFDLEtBQVIsQ0FBYyx1QkFBZCxFQUF1QyxHQUF2QyxFQUE0QyxLQUE1QyxFQUFBO09BUm1CO0lBQUEsQ0FOckIsQ0FBQTtBQUFBLElBaUJBLGVBQUEsR0FBa0Is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QSxPQUFPLENBQUMsR0FBUixDQUFZLFFBQVosRUFBc0IsSUFBdEIsRUFISTtRQUFBLENBRFIsQ0FBQSxDQURGO09BREE7YUFPQSxJQUFJLENBQUMsTUFSRTtJQUFBLENBOUdULENBQUE7QUFBQSxJQXlIQSxNQUFBLEdBQVMsU0FBQyxJQUFELEdBQUE7QUFDUCxNQUFBLElBQUcsYUFBYyxDQUFBLElBQUEsQ0FBZCxJQUF1QixFQUFBLGFBQWdCLENBQUEsSUFBQSxDQUFLLENBQUMsS0FBdEIsSUFBK0IsQ0FBekQ7QUFDRSxRQUFBLE1BQUEsQ0FBQSxhQUFxQixDQUFBLElBQUEsQ0FBckIsQ0FBQTtlQUNBLEdBQUEsQ0FBSSxRQUFKLEVBQWMsSUFBZCxDQUNFLENBQUMsSUFESCxDQUNRLFNBQUEsR0FBQTtpQkFBRyxPQUFPLENBQUMsR0FBUixDQUFZLFFBQVosRUFBc0IsSUFBdEIsRUFBSDtRQUFBLENBRFIsRUFGRjtPQURPO0lBQUEsQ0F6SFQsQ0FBQTtXQStIQTtBQUFBLE1BQUMsU0FBQSxPQUFEO0FBQUEsTUFBUyxNQUFBLElBQVQ7QUFBQSxNQUFjLE9BQUEsS0FBZDtBQUFBLE1BQW9CLEtBQUEsR0FBcEI7QUFBQSxNQUF3QixRQUFBLE1BQXhCO0FBQUEsTUFBK0IsUUFBQSxNQUEvQjtNQWhJbUI7RUFBQSxDQUFyQixDQUpBLENBQUE7QUFBQSIsInNvdXJjZXNDb250ZW50IjpbIm5nID0gYW5ndWxhci5tb2R1bGUgJ215QXBwJ1xuXG4jIFRoZSBcImplZWJ1c1wiIHNlcnZpY2UgYmVsb3cgaXMgdGhlIHNhbWUgZm9yIGFsbCBjbGllbnQtc2lkZSBhcHBsaWNhdGlvbnMuXG4jIEl0IGxldHMgYW5ndWxhciBjb25uZWN0IHRvIHRoZSBKZWVCdXMgc2VydmVyIGFuZCBzZW5kL3JlY2VpdmUgbWVzc2FnZXMuXG5uZy5mYWN0b3J5ICdqZWVidXMnLCAoJHJvb3RTY29wZSwgJHEpIC0+XG4gIHdzID0gbnVsbCAgICAgICAgICAjIHRoZSB3ZWJzb2NrZXQgb2JqZWN0LCB3aGlsZSBvcGVuXG4gIHNlcU51bSA9IDAgICAgICAgICAjIHVuaXF1ZSBzZXF1ZW5jZSBudW1iZXJzIGZvciBlYWNoIFJQQyByZXF1ZXN0XG4gIHJwY1Byb21pc2VzID0ge30gICAjIG1hcHMgc2VxTnVtIHRvIGEgcGVuZGluZyA8dGltZXJJZCxwcm9taXNlPiBlbnRyeVxuICB0cmFja2VkTW9kZWxzID0ge30gIyBrZWVwcyB0cmFjayBvZiB3aGljaCBwYXRocyBoYXZlIGJlZW4gYXR0YWNoZWRcblxuICAjIFVwZGF0ZSBvbmUgb3IgbW9yZSBvZiB0aGUgdHJhY2tlZCBtb2RlbHMgd2l0aCBhbiBpbmNvbWluZyBjaGFuZ2UuXG4gIHByb2Nlc3NNb2RlbFVwZGF0ZSA9IChrZXksIHZhbHVlKSAtPlxuICAgIGZvciBrLCBpbmZvIG9mIHRyYWNrZWRNb2RlbHNcbiAgICAgIGlmIGsgaXMga2V5LnNsaWNlKDAsIGsubGVuZ3RoKVxuICAgICAgICBzdWZmaXggPSBrZXkuc2xpY2Uoay5sZW5ndGgpXG4gICAgICAgIGlmIHZhbHVlXG4gICAgICAgICAgaW5mby5tb2RlbFtzdWZmaXhdID0gdmFsdWVcbiAgICAgICAgZWxzZVxuICAgICAgICAgIGRlbGV0ZSBpbmZvLm1vZGVsW3N1ZmZpeF1cbiAgICBjb25zb2xlLmVycm9yIFwic3B1cmlvdXMgbW9kZWwgdXBkYXRlXCIsIGtleSwgdmFsdWUgIHVubGVzcyBzdWZmaXhcblxuICAjIFJlc29sdmUgb3IgcmVqZWN0IGEgcGVuZGluZyBycGMgcHJvbWlzZS5cbiAgcHJvY2Vzc1JwY1JlcGx5ID0gKG4sIHJlc3VsdCwgZXJyKSAtPlxuICAgIGlmIHJwY1Byb21pc2VzW25dXG4gICAgICBbdGlkLGRdID0gcnBjUHJvbWlzZXNbbl1cbiAgICAgIGNsZWFyVGltZW91dCB0aWRcbiAgICAgIGlmIGVyclxuICAgICAgICBjb25zb2xlLmVycm9yIGVyclxuICAgICAgICBkLnJlamVjdCBlcnJcbiAgICAgIGVsc2VcbiAgICAgICAgZC5yZXNvbHZlIHJlc3VsdFxuICAgIGVsc2VcbiAgICAgIGNvbnNvbGUuZXJyb3IgXCJzcHVyaW91cyBycGMgcmVwbHlcIiwgbiwgcmVzdWx0LCBlcnJcblxuICAjIFNldCB1cCBhIHdlYnNvY2tldCBjb25uZWN0aW9uIHRvIHRoZSBKZWVCdXMgc2VydmVyLlxuICAjIFRoZSBhcHBUYWcgaXMgdGhlIGRlZmF1bHQgdGFnIHRvIHVzZSB3aGVuIHNlbmRpbmcgcmVxdWVzdHMgdG8gaXQuXG4gIGNvbm5lY3QgPSAoYXBwVGFnLCBwb3J0KSAtPlxuICAgIHBvcnQgPz0gbG9jYXRpb24ucG9ydCAjIHRoZSBkZWZhdWx0IHBvcnQgaXMgdGhlIHNhbWUgYXMgdGhlIEhUVFAgc2VydmVyXG5cbiAgICByZWNvbm5lY3QgPSAoZmlyc3RDYWxsKSAtPlxuICAgICAgIyB0aGUgd2Vic29ja2V0IGlzIHNlcnZlZCBmcm9tIHRoZSBzYW1lIHNpdGUgYXMgdGhlIHdlYiBwYWdlXG4gICAgICAjIHdzID0gbmV3IFdlYlNvY2tldCBcIndzOi8vI3tsb2NhdGlvbi5ob3N0fS93c1wiXG4gICAgICB3cyA9IG5ldyBXZWJTb2NrZXQgXCJ3czovLyN7bG9jYXRpb24uaG9zdG5hbWV9OiN7cG9ydH0vd3NcIiwgW2FwcFRhZ11cblxuICAgICAgd3Mub25vcGVuID0gLT5cbiAgICAgICAgIyBsb2NhdGlvbi5yZWxvYWQoKSAgdW5sZXNzIGZpcnN0Q2FsbFxuICAgICAgICBjb25zb2xlLmxvZyAnV1MgT3BlbidcblxuICAgICAgd3Mub25tZXNzYWdlID0gKG0pIC0+XG4gICAgICAgIGlmIG0uZGF0YSBpbnN0YW5jZW9mIEFycmF5QnVmZmVyXG4gICAgICAgICAgY29uc29sZS5sb2cgJ2JpbmFyeSBtc2cnLCBtXG4gICAgICAgICRyb290U2NvcGUuJGFwcGx5IC0+XG4gICAgICAgICAgZGF0YSA9IEpTT04ucGFyc2UobS5kYXRhKVxuICAgICAgICAgIGlmIG0uZGF0YVswXSBpcyAnWydcbiAgICAgICAgICAgIHByb2Nlc3NScGNSZXBseSBkYXRhLi4uXG4gICAgICAgICAgZWxzZVxuICAgICAgICAgICAgc3dpdGNoIGRhdGFcbiAgICAgICAgICAgICAgd2hlbiB0cnVlICMgcmVsb2FkIGFwcFxuICAgICAgICAgICAgICAgIHdpbmRvdy5sb2NhdGlvbi5yZWxvYWQgdHJ1ZVxuICAgICAgICAgICAgICB3aGVuIGZhbHNlICMgcmVmcmVzaCBzdHlsZXNoZWV0c1xuICAgICAgICAgICAgICAgIGNvbnNvbGUubG9nIFwiQ1NTIFJlbG9hZFwiXG4gICAgICAgICAgICAgICAgZm9yIGUgaW4gZG9jdW1lbnQuZ2V0RWxlbWVudHNCeVRhZ05hbWUgJ2xpbmsnXG4gICAgICAgICAgICAgICAgICBpZiBlLmhyZWYgYW5kIC9zdHlsZXNoZWV0L2kudGVzdCBlLnJlbFxuICAgICAgICAgICAgICAgICAgICBlLmhyZWYgPSBcIiN7ZS5ocmVmLnJlcGxhY2UgL1xcPy4qLywgJyd9PyN7RGF0ZS5ub3coKX1cIlxuICAgICAgICAgICAgICBlbHNlXG4gICAgICAgICAgICAgICAgIyBUT0RPOiBzaG91bGQgbm90IHdyaXRlIGludG8gdGhlIHJvb3Qgc2NvcGUgKG1lcmdlLCBwZXJoYXBzPylcbiAgICAgICAgICAgICAgICBmb3IgaywgdiBvZiBkYXRhXG4gICAgICAgICAgICAgICAgICAjICRyb290U2NvcGVba10gPSB2XG4gICAgICAgICAgICAgICAgICBwcm9jZXNzTW9kZWxVcGRhdGUgaywgdlxuXG4gICAgICAjIHdzLm9uZXJyb3IgPSAoZSkgLT5cbiAgICAgICMgICBjb25zb2xlLmxvZyAnRXJyb3InLCBlXG5cbiAgICAgIHdzLm9uY2xvc2UgPSAtPlxuICAgICAgICBjb25zb2xlLmxvZyAnV1MgQ2xvc2VkJ1xuICAgICAgICBzZXRUaW1lb3V0IHJlY29ubmVjdCwgMTAwMFxuXG4gICAgcmVjb25uZWN0IHRydWVcbiAgIFxuICAjIFNlbmQgYSBwYXlsb2FkIHRvIHRoZSBKZWVCdXMgc2VydmVyIG92ZXIgdGhlIHdlYnNvY2tldCBjb25uZWN0aW9uLlxuICAjIFRoZSBwYXlsb2FkIHNob3VsZCBiZSBhbiBvYmplY3QgKGFueXRoaW5nIGJ1dCBhcnJheSBpcyBzdXBwb3J0ZWQgZm9yIG5vdykuXG4gICMgVGhpcyBiZWNvbWVzIGFuIE1RVFQgbWVzc2FnZSB3aXRoIHRvcGljIFwic3YvPGFwcFRhZz4vaXAtPGFkZHI6cG9ydD5cIi5cbiAgc2VuZCA9IChwYXlsb2FkKSAtPlxuICAgIG1zZyA9IGFuZ3VsYXIudG9Kc29uIHBheWxvYWRcbiAgICBpZiBtc2dbMF0gaXMgJ1snXG4gICAgICBjb25zb2xlLmVycm9yIFwicGF5bG9hZCBjYW4ndCBiZSBhbiBhcnJheSAoI3twYXlsb2FkLmxlbmd0aH0gZWxlbWVudHMpXCJcbiAgICBlbHNlXG4gICAgICB3cy5zZW5kIG1zZ1xuICAgIEBcblxuICAjIFN0b3JlIGEga2V5L3ZhbHVlIHBhaXIgaW4gdGhlIEplZUJ1cyBkYXRhYmFzZSAoa2V5IG11c3Qgc3RhcnQgd2l0aCBcIi9cIikuXG4gIHN0b3JlID0gKGtleSwgdmFsdWUpIC0+XG4gICAgbXNnID0gYW5ndWxhci50b0pzb24gW2tleSwgdmFsdWVdXG4gICAgaWYgbXNnLnNsaWNlKDAsIDMpIGlzICdbXCIvJ1xuICAgICAgd3Muc2VuZCBtc2dcbiAgICBlbHNlXG4gICAgICBjb25zb2xlLmVycm9yICdrZXkgZG9lcyBub3Qgc3RhcnQgd2l0aCBcIi9cIjonLCBrZXlcbiAgICBAXG4gICAgICBcbiAgIyBQZXJmb3JtIGFuIFJQQyBjYWxsLCBpLmUuIHJlZ2lzdGVyIHJlc3VsdCBjYWxsYmFjayBhbmQgcmV0dXJuIGEgcHJvbWlzZS5cbiAgIyBUaGlzIGRvZXNuJ3QgdXNlIE1RVFQgdG8gYXZvaWQgYWRkaXRpb25hbCByb3VuZCB0cmlwcyBmb3IgZnJlcXVlbnQgY2FsbHMuXG4gIHJwYyA9IChhcmdzLi4uKSAtPlxuICAgIGQgPSAkcS5kZWZlcigpXG4gICAgbiA9ICsrc2VxTnVtXG4gICAgd3Muc2VuZCBhbmd1bGFyLnRvSnNvbiBbbiwgYXJncy4uLl1cbiAgICB0aWQgPSBzZXRUaW1lb3V0IC0+XG4gICAgICBjb25zb2xlLmVycm9yIFwiUlBDICN7bn06IG5vIHJlcG9uc2VcIiwgYXJnc1xuICAgICAgZGVsZXRlIHJwY1Byb21pc2VzW25dXG4gICAgICAkcm9vdFNjb3BlLiRhcHBseSAtPlxuICAgICAgICBkLnJlamVjdCgpXG4gICAgLCAxMDAwMCAjIDEwIHNlY29uZHMgc2hvdWxkIGJlIGVub3VnaCB0byBjb21wbGV0ZSBhbnkgcmVxdWVzdFxuICAgIHJwY1Byb21pc2VzW25dID0gW3RpZCwgZF1cbiAgICBkLnByb21pc2VcblxuICAjIEF0dGFjaCwgaS5lLiBnZXQgY29ycmVzcG9uZGluZyBkYXRhIGFzIGEgbW9kZWwgd2hpY2ggdHJhY2tzIGFsbCBjaGFuZ2VzLlxuICBhdHRhY2ggPSAocGF0aCkgLT5cbiAgICBpbmZvID0gdHJhY2tlZE1vZGVsc1twYXRoXSA/PSB7IG1vZGVsOiB7fSwgY291bnQ6IDAgfVxuICAgIGlmIGluZm8uY291bnQrKyBpcyAwXG4gICAgICBycGMgJ2F0dGFjaCcsIHBhdGhcbiAgICAgICAgLnRoZW4gKHIpIC0+XG4gICAgICAgICAgZm9yIGssIHYgb2YgclxuICAgICAgICAgICAgcHJvY2Vzc01vZGVsVXBkYXRlIGssIHZcbiAgICAgICAgICBjb25zb2xlLmxvZyAnYXR0YWNoJywgcGF0aFxuICAgIGluZm8ubW9kZWxcblxuICAjIFVuZG8gdGhlIGVmZmVjdHMgb2YgYXR0YWNoaW5nLCBpLmUuIHN0b3AgZm9sbG93aW5nIGNoYW5nZXMuXG4gIGRldGFjaCA9IChwYXRoKSAtPlxuICAgIGlmIHRyYWNrZWRNb2RlbHNbcGF0aF0gJiYgLS10cmFja2VkTW9kZWxzW3BhdGhdLmNvdW50IDw9IDBcbiAgICAgIGRlbGV0ZSB0cmFja2VkTW9kZWxzW3BhdGhdXG4gICAgICBycGMgJ2RldGFjaCcsIHBhdGhcbiAgICAgICAgLnRoZW4gLT4gY29uc29sZS5sb2cgJ2RldGFjaCcsIHBhdGhcblxuICB7Y29ubmVjdCxzZW5kLHN0b3JlLHJwYyxhdHRhY2gsZGV0YWNofVxuIl19
