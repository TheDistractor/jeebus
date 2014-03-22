(function() {
  var ng,
    __slice = [].slice;

  ng = angular.module('myApp');

  console.log('NG', angular.version.full);

  ng.config(function($urlRouterProvider, $locationProvider) {
    $urlRouterProvider.otherwise('/');
    return $locationProvider.html5Mode(true);
  });

  ng.factory('jeebus', function($rootScope, $q) {
    var attach, connect, detach, gadget, get, processModelUpdate, processRpcReply, put, rpc, rpcPromises, send, seqNum, trackedModels, ws;
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
    processRpcReply = function(reply) {
      var d, e, msg, n, result, t, _ref;
      n = reply[0], msg = reply[1], result = reply[2];
      _ref = rpcPromises[n], t = _ref[0], d = _ref[1], e = _ref[2];
      if (d) {
        clearTimeout(t);
        if (msg === true) {
          rpcPromises[n][1] = null;
          return d.resolve(function(ee) {
            return rpcPromises[n][2] = ee;
          });
        } else if (msg === "" && reply.length === 3) {
          return d.resolve(result);
        } else if (msg !== "" && reply.length === 2) {
          console.error(msg);
          return d.reject(msg);
        } else {
          return console.error.apply(console, ["bad rpc reply"].concat(__slice.call(reply)));
        }
      } else if (e) {
        if (msg === false) {
          return delete rpcPromises[n];
        } else if (msg !== "" && reply.length > 2) {
          return e.emit(msg, reply.slice(2));
        } else {
          return console.error.apply(console, ["bad rpc event"].concat(__slice.call(reply)));
        }
      } else {
        return console.error.apply(console, ["spurious rpc reply"].concat(__slice.call(reply)));
      }
    };
    connect = function(appTag) {
      var reconnect;
      reconnect = function(firstCall) {
        ws = new WebSocket("ws://" + location.host + "/ws", [appTag]);
        ws.onopen = function() {
          console.log('WS Open');
          return $rootScope.$apply(function() {
            return $rootScope.serverStatus = 'connected';
          });
        };
        ws.onmessage = function(m) {
          if (m.data instanceof ArrayBuffer) {
            console.log('binary msg', m);
          }
          return $rootScope.$apply(function() {
            var data, e, k, v, _i, _len, _ref, _results, _results1;
            data = JSON.parse(m.data);
            switch (typeof data) {
              case 'object':
                if (Array.isArray(data)) {
                  return processRpcReply(data);
                } else {
                  _results = [];
                  for (k in data) {
                    v = data[k];
                    _results.push(processModelUpdate(k, v));
                  }
                  return _results;
                }
                break;
              case 'boolean':
                if (data) {
                  return window.location.reload(true);
                } else {
                  console.log("CSS Reload");
                  _ref = document.getElementsByTagName('link');
                  _results1 = [];
                  for (_i = 0, _len = _ref.length; _i < _len; _i++) {
                    e = _ref[_i];
                    if (e.href && /stylesheet/i.test(e.rel)) {
                      _results1.push(e.href = "" + (e.href.replace(/\?.*/, '')) + "?" + (Date.now()));
                    } else {
                      _results1.push(void 0);
                    }
                  }
                  return _results1;
                }
                break;
              default:
                return console.log('Server msg:', data);
            }
          });
        };
        return ws.onclose = function() {
          console.log('WS Closed');
          $rootScope.$apply(function() {
            return $rootScope.serverStatus = 'disconnected';
          });
          return setTimeout(reconnect, 1000);
        };
      };
      return reconnect(true);
    };
    send = function(payload) {
      ws.send(angular.toJson(payload));
      return this;
    };
    get = function(key) {
      return rpc('get', key);
    };
    put = function(key, value) {
      send([0, 'put', key, value]);
      return this;
    };
    rpc = function() {
      var args, cmd, d, n, t;
      cmd = arguments[0], args = 2 <= arguments.length ? __slice.call(arguments, 1) : [];
      d = $q.defer();
      n = ++seqNum;
      ws.send(angular.toJson([cmd, n].concat(__slice.call(args))));
      t = setTimeout(function() {
        console.error("RPC " + n + ": no reponse", args);
        delete rpcPromises[n];
        return $rootScope.$apply(function() {
          return d.reject();
        });
      }, 10000);
      rpcPromises[n] = [t, d, null];
      return d.promise;
    };
    gadget = function() {
      var args, e;
      args = 1 <= arguments.length ? __slice.call(arguments, 0) : [];
      e = new EventEmitter;
      rpc.apply(null, args).then(function(eeSetter) {
        return eeSetter(e);
      });
      return e;
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
        rpc('detach', path).then(function() {
          return console.log('detach', path);
        });
      }
      return this;
    };
    window.send = send;
    return {
      connect: connect,
      send: send,
      get: get,
      put: put,
      rpc: rpc,
      gadget: gadget,
      attach: attach,
      detach: detach
    };
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiamVlYnVzLmNvZmZlZSJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUFBLE1BQUEsRUFBQTtJQUFBLGtCQUFBOztBQUFBLEVBQUEsRUFBQSxHQUFLLE9BQU8sQ0FBQyxNQUFSLENBQWUsT0FBZixDQUFMLENBQUE7O0FBQUEsRUFFQSxPQUFPLENBQUMsR0FBUixDQUFZLElBQVosRUFBa0IsT0FBTyxDQUFDLE9BQU8sQ0FBQyxJQUFsQyxDQUZBLENBQUE7O0FBQUEsRUFJQSxFQUFFLENBQUMsTUFBSCxDQUFVLFNBQUMsa0JBQUQsRUFBcUIsaUJBQXJCLEdBQUE7QUFDUixJQUFBLGtCQUFrQixDQUFDLFNBQW5CLENBQTZCLEdBQTdCLENBQUEsQ0FBQTtXQUNBLGlCQUFpQixDQUFDLFNBQWxCLENBQTRCLElBQTVCLEVBRlE7RUFBQSxDQUFWLENBSkEsQ0FBQTs7QUFBQSxFQVVBLEVBQUUsQ0FBQyxPQUFILENBQVcsUUFBWCxFQUFxQixTQUFDLFVBQUQsRUFBYSxFQUFiLEdBQUE7QUFDbkIsUUFBQSxpSUFBQTtBQUFBLElBQUEsRUFBQSxHQUFLLElBQUwsQ0FBQTtBQUFBLElBQ0EsTUFBQSxHQUFTLENBRFQsQ0FBQTtBQUFBLElBRUEsV0FBQSxHQUFjLEVBRmQsQ0FBQTtBQUFBLElBR0EsYUFBQSxHQUFnQixFQUhoQixDQUFBO0FBQUEsSUFNQSxrQkFBQSxHQUFxQixTQUFDLEdBQUQsRUFBTSxLQUFOLEdBQUE7QUFDbkIsVUFBQSxlQUFBO0FBQUEsV0FBQSxrQkFBQTtnQ0FBQTtBQUNFLFFBQUEsSUFBRyxDQUFBLEtBQUssR0FBRyxDQUFDLEtBQUosQ0FBVSxDQUFWLEVBQWEsQ0FBQyxDQUFDLE1BQWYsQ0FBUjtBQUNFLFVBQUEsTUFBQSxHQUFTLEdBQUcsQ0FBQyxLQUFKLENBQVUsQ0FBQyxDQUFDLE1BQVosQ0FBVCxDQUFBO0FBQ0EsVUFBQSxJQUFHLEtBQUg7QUFDRSxZQUFBLElBQUksQ0FBQyxLQUFN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O0VBQUEsQ0FBckIsQ0FWQSxDQUFBO0FBQUEiLCJzb3VyY2VzQ29udGVudCI6WyJuZyA9IGFuZ3VsYXIubW9kdWxlICdteUFwcCdcblxuY29uc29sZS5sb2cgJ05HJywgYW5ndWxhci52ZXJzaW9uLmZ1bGxcblxubmcuY29uZmlnICgkdXJsUm91dGVyUHJvdmlkZXIsICRsb2NhdGlvblByb3ZpZGVyKSAtPlxuICAkdXJsUm91dGVyUHJvdmlkZXIub3RoZXJ3aXNlICcvJ1xuICAkbG9jYXRpb25Qcm92aWRlci5odG1sNU1vZGUgdHJ1ZVxuICBcbiMgVGhlIFwiamVlYnVzXCIgc2VydmljZSBiZWxvdyBpcyB0aGUgc2FtZSBmb3IgYWxsIGNsaWVudC1zaWRlIGFwcGxpY2F0aW9ucy5cbiMgSXQgbGV0cyBhbmd1bGFyIGNvbm5lY3QgdG8gdGhlIEplZUJ1cyBzZXJ2ZXIgYW5kIHNlbmQvcmVjZWl2ZSBtZXNzYWdlcy5cbm5nLmZhY3RvcnkgJ2plZWJ1cycsICgkcm9vdFNjb3BlLCAkcSkgLT5cbiAgd3MgPSBudWxsICAgICAgICAgICMgdGhlIHdlYnNvY2tldCBvYmplY3QsIHdoaWxlIG9wZW5cbiAgc2VxTnVtID0gMCAgICAgICAgICMgdW5pcXVlIHNlcXVlbmNlIG51bWJlcnMgZm9yIGVhY2ggUlBDIHJlcXVlc3RcbiAgcnBjUHJvbWlzZXMgPSB7fSAgICMgbWFwcyBzZXFOdW0gdG8gYSBwZW5kaW5nIDx0aW1lcklkLHByb21pc2UsZW1pdHRlcj4gZW50cnlcbiAgdHJhY2tlZE1vZGVscyA9IHt9ICMga2VlcHMgdHJhY2sgb2Ygd2hpY2ggcGF0aHMgaGF2ZSBiZWVuIGF0dGFjaGVkXG5cbiAgIyBVcGRhdGUgb25lIG9yIG1vcmUgb2YgdGhlIHRyYWNrZWQgbW9kZWxzIHdpdGggYW4gaW5jb21pbmcgY2hhbmdlLlxuICBwcm9jZXNzTW9kZWxVcGRhdGUgPSAoa2V5LCB2YWx1ZSkgLT5cbiAgICBmb3IgaywgaW5mbyBvZiB0cmFja2VkTW9kZWxzXG4gICAgICBpZiBrIGlzIGtleS5zbGljZSgwLCBrLmxlbmd0aClcbiAgICAgICAgc3VmZml4ID0ga2V5LnNsaWNlKGsubGVuZ3RoKVxuICAgICAgICBpZiB2YWx1ZVxuICAgICAgICAgIGluZm8ubW9kZWxbc3VmZml4XSA9IHZhbHVlXG4gICAgICAgIGVsc2VcbiAgICAgICAgICBkZWxldGUgaW5mby5tb2RlbFtzdWZmaXhdXG4gICAgY29uc29sZS5lcnJvciBcInNwdXJpb3VzIG1vZGVsIHVwZGF0ZVwiLCBrZXksIHZhbHVlICB1bmxlc3Mgc3VmZml4XG5cbiAgIyBSZXNvbHZlIG9yIHJlamVjdCBhIHBlbmRpbmcgcnBjIHByb21pc2UuXG4gIHByb2Nlc3NScGNSZXBseSA9IChyZXBseSkgLT5cbiAgICBbbixtc2cscmVzdWx0XSA9IHJlcGx5XG4gICAgW3QsZCxlXSA9IHJwY1Byb21pc2VzW25dXG4gICAgaWYgZFxuICAgICAgY2xlYXJUaW1lb3V0IHRcbiAgICAgIGlmIG1zZyBpcyB0cnVlICMgc3RhcnQgc3RyZWFtaW5nXG4gICAgICAgIHJwY1Byb21pc2VzW25dWzFdID0gbnVsbFxuICAgICAgICBkLnJlc29sdmUgKGVlKSAtPlxuICAgICAgICAgIHJwY1Byb21pc2VzW25dWzJdID0gZWVcbiAgICAgIGVsc2UgaWYgbXNnIGlzIFwiXCIgYW5kIHJlcGx5Lmxlbmd0aCBpcyAzXG4gICAgICAgIGQucmVzb2x2ZSByZXN1bHRcbiAgICAgIGVsc2UgaWYgbXNnIGlzbnQgXCJcIiBhbmQgcmVwbHkubGVuZ3RoIGlzIDJcbiAgICAgICAgY29uc29sZS5lcnJvciBtc2dcbiAgICAgICAgZC5yZWplY3QgbXNnXG4gICAgICBlbHNlXG4gICAgICAgIGNvbnNvbGUuZXJyb3IgXCJiYWQgcnBjIHJlcGx5XCIsIHJlcGx5Li4uXG4gICAgZWxzZSBpZiBlXG4gICAgICBpZiBtc2cgaXMgZmFsc2UgIyBzdG9wIHN0cmVhbWluZ1xuICAgICAgICBkZWxldGUgcnBjUHJvbWlzZXNbbl1cbiAgICAgIGVsc2UgaWYgbXNnIGlzbnQgXCJcIiBhbmQgcmVwbHkubGVuZ3RoID4gMlxuICAgICAgICBlLmVtaXQgbXNnLCByZXBseS5zbGljZSgyKVxuICAgICAgZWxzZVxuICAgICAgICBjb25zb2xlLmVycm9yIFwiYmFkIHJwYyBldmVudFwiLCByZXBseS4uLlxuICAgIGVsc2VcbiAgICAgIGNvbnNvbGUuZXJyb3IgXCJzcHVyaW91cyBycGMgcmVwbHlcIiwgcmVwbHkuLi5cblxuICAjIFNldCB1cCBhIHdlYnNvY2tldCBjb25uZWN0aW9uIHRvIHRoZSBKZWVCdXMgc2VydmVyLlxuICAjIFRoZSBhcHBUYWcgaXMgdGhlIGRlZmF1bHQgdGFnIHRvIHVzZSB3aGVuIHNlbmRpbmcgcmVxdWVzdHMgdG8gaXQuXG4gIGNvbm5lY3QgPSAoYXBwVGFnKSAtPlxuXG4gICAgcmVjb25uZWN0ID0gKGZpcnN0Q2FsbCkgLT5cbiAgICAgICMgdGhlIHdlYnNvY2tldCBpcyBzZXJ2ZWQgZnJvbSB0aGUgc2FtZSBzaXRlIGFzIHRoZSB3ZWIgcGFnZVxuICAgICAgd3MgPSBuZXcgV2ViU29ja2V0IFwid3M6Ly8je2xvY2F0aW9uLmhvc3R9L3dzXCIsIFthcHBUYWddXG5cbiAgICAgIHdzLm9ub3BlbiA9IC0+XG4gICAgICAgICMgbG9jYXRpb24ucmVsb2FkKCkgIHVubGVzcyBmaXJzdENhbGxcbiAgICAgICAgY29uc29sZS5sb2cgJ1dTIE9wZW4nXG4gICAgICAgICRyb290U2NvcGUuJGFwcGx5IC0+XG4gICAgICAgICAgJHJvb3RTY29wZS5zZXJ2ZXJTdGF0dXMgPSAnY29ubmVjdGVkJ1xuXG4gICAgICB3cy5vbm1lc3NhZ2UgPSAobSkgLT5cbiAgICAgICAgaWYgbS5kYXRhIGluc3RhbmNlb2YgQXJyYXlCdWZmZXJcbiAgICAgICAgICBjb25zb2xlLmxvZyAnYmluYXJ5IG1zZycsIG1cbiAgICAgICAgJHJvb3RTY29wZS4kYXBwbHkgLT5cbiAgICAgICAgICBkYXRhID0gSlNPTi5wYXJzZSBtLmRhdGFcbiAgICAgICAgICBzd2l0Y2ggdHlwZW9mIGRhdGFcbiAgICAgICAgICAgIHdoZW4gJ29iamVjdCdcbiAgICAgICAgICAgICAgaWYgQXJyYXkuaXNBcnJheSBkYXRhXG4gICAgICAgICAgICAgICAgcHJvY2Vzc1JwY1JlcGx5IGRhdGFcbiAgICAgICAgICAgICAgZWxzZVxuICAgICAgICAgICAgICAgIGZvciBrLCB2IG9mIGRhdGFcbiAgICAgICAgICAgICAgICAgIHByb2Nlc3NNb2RlbFVwZGF0ZSBrLCB2XG4gICAgICAgICAgICB3aGVuICdib29sZWFuJ1xuICAgICAgICAgICAgICBpZiBkYXRhICMgcmVsb2FkIGFwcFxuICAgICAgICAgICAgICAgIHdpbmRvdy5sb2NhdGlvbi5yZWxvYWQgdHJ1ZVxuICAgICAgICAgICAgICBlbHNlICMgcmVmcmVzaCBzdHlsZXNoZWV0c1xuICAgICAgICAgICAgICAgIGNvbnNvbGUubG9nIFwiQ1NTIFJlbG9hZFwiXG4gICAgICAgICAgICAgICAgZm9yIGUgaW4gZG9jdW1lbnQuZ2V0RWxlbWVudHNCeVRhZ05hbWUgJ2xpbmsnXG4gICAgICAgICAgICAgICAgICBpZiBlLmhyZWYgYW5kIC9zdHlsZXNoZWV0L2kudGVzdCBlLnJlbFxuICAgICAgICAgICAgICAgICAgICBlLmhyZWYgPSBcIiN7ZS5ocmVmLnJlcGxhY2UgL1xcPy4qLywgJyd9PyN7RGF0ZS5ub3coKX1cIlxuICAgICAgICAgICAgZWxzZVxuICAgICAgICAgICAgICBjb25zb2xlLmxvZyAnU2VydmVyIG1zZzonLCBkYXRhXG5cbiAgICAgICMgd3Mub25lcnJvciA9IChlKSAtPlxuICAgICAgIyAgIGNvbnNvbGUubG9nICdFcnJvcicsIGVcblxuICAgICAgd3Mub25jbG9zZSA9IC0+XG4gICAgICAgIGNvbnNvbGUubG9nICdXUyBDbG9zZWQnXG4gICAgICAgICRyb290U2NvcGUuJGFwcGx5IC0+XG4gICAgICAgICAgJHJvb3RTY29wZS5zZXJ2ZXJTdGF0dXMgPSAnZGlzY29ubmVjdGVkJ1xuICAgICAgICBzZXRUaW1lb3V0IHJlY29ubmVjdCwgMTAwMFxuXG4gICAgcmVjb25uZWN0IHRydWVcbiAgIFxuICAjIFNlbmQgYSBwYXlsb2FkIHRvIHRoZSBKZWVCdXMgc2VydmVyIG92ZXIgdGhlIHdlYnNvY2tldCBjb25uZWN0aW9uLlxuICAjIFRoZSBwYXlsb2FkIHNob3VsZCBiZSBhbiBvYmplY3QgKGFueXRoaW5nIGJ1dCBhcnJheSBpcyBzdXBwb3J0ZWQgZm9yIG5vdykuXG4gIHNlbmQgPSAocGF5bG9hZCkgLT5cbiAgICB3cy5zZW5kIGFuZ3VsYXIudG9Kc29uIHBheWxvYWRcbiAgICBAXG5cbiAgIyBGZXRjaCBhIGtleS92YWx1ZSBwYWlyIGZyb20gdGhlIHNlcnZlciBkYXRhYmFzZSwgdmFsdWUgcmV0dXJuZWQgYXMgcHJvbWlzZS5cbiAgZ2V0ID0gKGtleSkgLT5cbiAgICBycGMgJ2dldCcsIGtleVxuICAgICAgXG4gICMgU3RvcmUgYSBrZXkvdmFsdWUgcGFpciBpbiB0aGUgc2VydmVyIGRhdGFiYXNlLlxuICBwdXQgPSAoa2V5LCB2YWx1ZSkgLT5cbiAgICBzZW5kIFswLCAncHV0Jywga2V5LCB2YWx1ZV1cbiAgICBAXG4gICAgICBcbiAgIyBQZXJmb3JtIGFuIFJQQyBjYWxsLCBpLmUuIHJlZ2lzdGVyIHJlc3VsdCBjYWxsYmFjayBhbmQgcmV0dXJuIGEgcHJvbWlzZS5cbiAgcnBjID0gKGNtZCwgYXJncy4uLikgLT5cbiAgICBkID0gJHEuZGVmZXIoKVxuICAgIG4gPSArK3NlcU51bVxuICAgIHdzLnNlbmQgYW5ndWxhci50b0pzb24gW2NtZCwgbiwgYXJncy4uLl1cbiAgICB0ID0gc2V0VGltZW91dCAtPlxuICAgICAgY29uc29sZS5lcnJvciBcIlJQQyAje259OiBubyByZXBvbnNlXCIsIGFyZ3NcbiAgICAgIGRlbGV0ZSBycGNQcm9taXNlc1tuXVxuICAgICAgJHJvb3RTY29wZS4kYXBwbHkgLT5cbiAgICAgICAgZC5yZWplY3QoKVxuICAgICwgMTAwMDAgIyAxMCBzZWNvbmRzIHNob3VsZCBiZSBlbm91Z2ggdG8gY29tcGxldGUgYW55IHJlcXVlc3RcbiAgICBycGNQcm9taXNlc1tuXSA9IFt0LCBkLCBudWxsXVxuICAgIGQucHJvbWlzZVxuXG4gICMgTGF1bmNoIGEgZ2FkZ2V0IG9uIHRoZSBzZXJ2ZXIgYW5kIHJldHVybiBpdHMgcmVzdWx0cyB2aWEgZXZlbnRzLlxuICBnYWRnZXQgPSAoYXJncy4uLikgLT5cbiAgICBlID0gbmV3IEV2ZW50RW1pdHRlclxuICAgIHJwYyBhcmdzLi4uXG4gICAgICAudGhlbiAoZWVTZXR0ZXIpIC0+XG4gICAgICAgIGVlU2V0dGVyIGVcbiAgICBlXG4gIFxuICAjIEF0dGFjaCwgaS5lLiBnZXQgY29ycmVzcG9uZGluZyBkYXRhIGFzIGEgbW9kZWwgd2hpY2ggdHJhY2tzIGFsbCBjaGFuZ2VzLlxuICBhdHRhY2ggPSAocGF0aCkgLT5cbiAgICBpbmZvID0gdHJhY2tlZE1vZGVsc1twYXRoXSA/PSB7IG1vZGVsOiB7fSwgY291bnQ6IDAgfVxuICAgIGlmIGluZm8uY291bnQrKyBpcyAwXG4gICAgICBycGMgJ2F0dGFjaCcsIHBhdGhcbiAgICAgICAgLnRoZW4gKHIpIC0+XG4gICAgICAgICAgZm9yIGssIHYgb2YgclxuICAgICAgICAgICAgcHJvY2Vzc01vZGVsVXBkYXRlIGssIHZcbiAgICAgICAgICBjb25zb2xlLmxvZyAnYXR0YWNoJywgcGF0aFxuICAgIGluZm8ubW9kZWxcblxuICAjIFVuZG8gdGhlIGVmZmVjdHMgb2YgYXR0YWNoaW5nLCBpLmUuIHN0b3AgZm9sbG93aW5nIGNoYW5nZXMuXG4gIGRldGFjaCA9IChwYXRoKSAtPlxuICAgIGlmIHRyYWNrZWRNb2RlbHNbcGF0aF0gJiYgLS10cmFja2VkTW9kZWxzW3BhdGhdLmNvdW50IDw9IDBcbiAgICAgIGRlbGV0ZSB0cmFja2VkTW9kZWxzW3BhdGhdXG4gICAgICBycGMgJ2RldGFjaCcsIHBhdGhcbiAgICAgICAgLnRoZW4gLT4gY29uc29sZS5sb2cgJ2RldGFjaCcsIHBhdGhcbiAgICBAXG5cbiAgd2luZG93LnNlbmQgPSBzZW5kICMgY29uc29sZSBhY2Nlc3MsIGZvciBkZWJ1Z2dpbmdcbiAge2Nvbm5lY3Qsc2VuZCxnZXQscHV0LHJwYyxnYWRnZXQsYXR0YWNoLGRldGFjaH1cbiJdfQ==
