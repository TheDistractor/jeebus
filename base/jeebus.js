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
    var connect, gadget, get, processRpcReply, put, rpc, rpcPromises, send, seqNum, ws;
    ws = null;
    seqNum = 0;
    rpcPromises = {};
    processRpcReply = function() {
      var deferred, emitter, msg, n, reply, timer, _ref;
      n = arguments[0], msg = arguments[1], reply = 3 <= arguments.length ? __slice.call(arguments, 2) : [];
      _ref = rpcPromises[n], timer = _ref.timer, deferred = _ref.deferred, emitter = _ref.emitter;
      if (deferred) {
        clearTimeout(timer);
        if (msg === true) {
          rpcPromises[n].deferred = null;
          deferred.resolve(function(ee) {
            return rpcPromises[n].emitter = ee;
          });
          return;
        }
        if (msg === "" && reply.length) {
          deferred.resolve(reply[0]);
        } else if (msg && reply.length === 0) {
          console.error(msg);
          deferred.reject(msg);
        } else {
          console.error.apply(console, ["bad rpc reply", n, msg].concat(__slice.call(reply)));
        }
        return delete rpcPromises[n];
      } else if (emitter) {
        if (msg && reply.length) {
          return emitter.emit(msg, reply[0]);
        } else {
          delete rpcPromises[n];
          return emitter.emit('close', reply[0]);
        }
      } else {
        return console.error.apply(console, ["spurious rpc reply", n, msg].concat(__slice.call(reply)));
      }
    };
    connect = function(appTag) {
      var reconnect;
      reconnect = function(firstCall) {
        ws = new WebSocket("ws://" + location.host + "/ws", [appTag]);
        ws.onopen = function() {
          console.log('WS Open');
          return $rootScope.$apply(function() {
            return $rootScope.$broadcast('ws-open');
          });
        };
        ws.onmessage = function(m) {
          if (m.data instanceof ArrayBuffer) {
            console.log('binary msg', m);
          }
          return $rootScope.$apply(function() {
            var data, e, _i, _len, _ref, _results;
            data = JSON.parse(m.data);
            switch (typeof data) {
              case 'object':
                if (Array.isArray(data)) {
                  return processRpcReply.apply(null, data);
                } else {
                  return console.log({
                    "spurious object received": m
                  });
                }
                break;
              case 'boolean':
                if (data) {
                  return window.location.reload(true);
                } else {
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
                }
                break;
              default:
                return console.log('Server msg:', data);
            }
          });
        };
        return ws.onclose = function() {
          console.log('WS Lost');
          $rootScope.$apply(function() {
            return $rootScope.$broadcast('ws-lost');
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
      rpcPromises[n] = {
        timer: t,
        deferred: d
      };
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
    window.send = send;
    return {
      connect: connect,
      send: send,
      get: get,
      put: put,
      rpc: rpc,
      gadget: gadget
    };
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiamVlYnVzLmNvZmZlZSJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUFBLE1BQUEsRUFBQTtJQUFBLGtCQUFBOztBQUFBLEVBQUEsRUFBQSxHQUFLLE9BQU8sQ0FBQyxNQUFSLENBQWUsT0FBZixDQUFMLENBQUE7O0FBQUEsRUFFQSxPQUFPLENBQUMsR0FBUixDQUFZLElBQVosRUFBa0IsT0FBTyxDQUFDLE9BQU8sQ0FBQyxJQUFsQyxDQUZBLENBQUE7O0FBQUEsRUFJQSxFQUFFLENBQUMsTUFBSCxDQUFVLFNBQUMsa0JBQUQsRUFBcUIsaUJBQXJCLEdBQUE7QUFDUixJQUFBLGtCQUFrQixDQUFDLFNBQW5CLENBQTZCLEdBQTdCLENBQUEsQ0FBQTtXQUNBLGlCQUFpQixDQUFDLFNBQWxCLENBQTRCLElBQTVCLEVBRlE7RUFBQSxDQUFWLENBSkEsQ0FBQTs7QUFBQSxFQVVBLEVBQUUsQ0FBQyxPQUFILENBQVcsUUFBWCxFQUFxQixTQUFDLFVBQUQsRUFBYSxFQUFiLEdBQUE7QUFDbkIsUUFBQSw4RUFBQTtBQUFBLElBQUEsRUFBQSxHQUFLLElBQUwsQ0FBQTtBQUFBLElBQ0EsTUFBQSxHQUFTLENBRFQsQ0FBQTtBQUFBLElBRUEsV0FBQSxHQUFjLEVBRmQsQ0FBQTtBQUFBLElBS0EsZUFBQSxHQUFrQixTQUFBLEdBQUE7QUFDaEIsVUFBQSw2Q0FBQTtBQUFBLE1BRGlCLGtCQUFHLG9CQUFLLCtEQUN6QixDQUFBO0FBQUEsTUFBQSxPQUEyQixXQUFZLENBQUEsQ0FBQSxDQUF2QyxFQUFDLGFBQUEsS0FBRCxFQUFPLGdCQUFBLFFBQVAsRUFBZ0IsZUFBQSxPQUFoQixDQUFBO0FBQ0EsTUFBQSxJQUFHLFFBQUg7QUFDRSxRQUFBLFlBQUEsQ0FBYSxLQUFiLENBQUEsQ0FBQTtBQUNBLFFBQUEsSUFBRyxHQUFBLEtBQU8sSUFBVjtBQUNFLFVBQUEsV0FBWSxDQUFBLENBQUEsQ0FBRSxDQUFDLFFBQWYsR0FB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QSxRQUFVLFFBQUEsRUFBVSxDQUFwQjtPQVRqQixDQUFBO2FBVUEsQ0FBQyxDQUFDLFFBWEU7SUFBQSxDQTlGTixDQUFBO0FBQUEsSUE0R0EsTUFBQSxHQUFTLFNBQUEsR0FBQTtBQUNQLFVBQUEsT0FBQTtBQUFBLE1BRFEsOERBQ1IsQ0FBQTtBQUFBLE1BQUEsQ0FBQSxHQUFJLEdBQUEsQ0FBQSxZQUFKLENBQUE7QUFBQSxNQUNBLEdBQUEsYUFBSSxJQUFKLENBQ0UsQ0FBQyxJQURILENBQ1EsU0FBQyxRQUFELEdBQUE7ZUFDSixRQUFBLENBQVMsQ0FBVCxFQURJO01BQUEsQ0FEUixDQURBLENBQUE7YUFJQSxFQUxPO0lBQUEsQ0E1R1QsQ0FBQTtBQUFBLElBbUhBLE1BQU0sQ0FBQyxJQUFQLEdBQWMsSUFuSGQsQ0FBQTtXQW9IQTtBQUFBLE1BQUMsU0FBQSxPQUFEO0FBQUEsTUFBUyxNQUFBLElBQVQ7QUFBQSxNQUFjLEtBQUEsR0FBZDtBQUFBLE1BQWtCLEtBQUEsR0FBbEI7QUFBQSxNQUFzQixLQUFBLEdBQXRCO0FBQUEsTUFBMEIsUUFBQSxNQUExQjtNQXJIbUI7RUFBQSxDQUFyQixDQVZBLENBQUE7QUFBQSIsInNvdXJjZXNDb250ZW50IjpbIm5nID0gYW5ndWxhci5tb2R1bGUgJ215QXBwJ1xuXG5jb25zb2xlLmxvZyAnTkcnLCBhbmd1bGFyLnZlcnNpb24uZnVsbFxuXG5uZy5jb25maWcgKCR1cmxSb3V0ZXJQcm92aWRlciwgJGxvY2F0aW9uUHJvdmlkZXIpIC0+XG4gICR1cmxSb3V0ZXJQcm92aWRlci5vdGhlcndpc2UgJy8nXG4gICRsb2NhdGlvblByb3ZpZGVyLmh0bWw1TW9kZSB0cnVlXG4gIFxuIyBUaGUgXCJqZWVidXNcIiBzZXJ2aWNlIGJlbG93IGlzIHRoZSBzYW1lIGZvciBhbGwgY2xpZW50LXNpZGUgYXBwbGljYXRpb25zLlxuIyBJdCBsZXRzIGFuZ3VsYXIgY29ubmVjdCB0byB0aGUgSmVlQnVzIHNlcnZlciBhbmQgc2VuZC9yZWNlaXZlIG1lc3NhZ2VzLlxubmcuZmFjdG9yeSAnamVlYnVzJywgKCRyb290U2NvcGUsICRxKSAtPlxuICB3cyA9IG51bGwgICAgICAgICAgIyB0aGUgd2Vic29ja2V0IG9iamVjdCwgd2hpbGUgb3BlblxuICBzZXFOdW0gPSAwICAgICAgICAgIyB1bmlxdWUgc2VxdWVuY2UgbnVtYmVycyBmb3IgZWFjaCBSUEMgcmVxdWVzdFxuICBycGNQcm9taXNlcyA9IHt9ICAgIyBtYXBzIHNlcU51bSB0byBhIHBlbmRpbmcge3RpbWVyLGRlZmVycmVkLGVtaXR0ZXJ9IGVudHJ5XG5cbiAgIyBSZXNvbHZlIG9yIHJlamVjdCBhIHBlbmRpbmcgcnBjIHByb21pc2UuIEFsc28gaGFuZGxlIHN0cmVhbWVkIHJlc3VsdHMuXG4gIHByb2Nlc3NScGNSZXBseSA9IChuLCBtc2csIHJlcGx5Li4uKSAtPlxuICAgIHt0aW1lcixkZWZlcnJlZCxlbWl0dGVyfSA9IHJwY1Byb21pc2VzW25dXG4gICAgaWYgZGVmZXJyZWRcbiAgICAgIGNsZWFyVGltZW91dCB0aW1lclxuICAgICAgaWYgbXNnIGlzIHRydWUgIyBzdGFydCBzdHJlYW1pbmdcbiAgICAgICAgcnBjUHJvbWlzZXNbbl0uZGVmZXJyZWQgPSBudWxsXG4gICAgICAgIGRlZmVycmVkLnJlc29sdmUgKGVlKSAtPlxuICAgICAgICAgIHJwY1Byb21pc2VzW25dLmVtaXR0ZXIgPSBlZVxuICAgICAgICByZXR1cm5cbiAgICAgIGlmIG1zZyBpcyBcIlwiIGFuZCByZXBseS5sZW5ndGhcbiAgICAgICAgZGVmZXJyZWQucmVzb2x2ZSByZXBseVswXVxuICAgICAgZWxzZSBpZiBtc2cgYW5kIHJlcGx5Lmxlbmd0aCA9PSAwXG4gICAgICAgIGNvbnNvbGUuZXJyb3IgbXNnXG4gICAgICAgIGRlZmVycmVkLnJlamVjdCBtc2dcbiAgICAgIGVsc2VcbiAgICAgICAgY29uc29sZS5lcnJvciBcImJhZCBycGMgcmVwbHlcIiwgbiwgbXNnLCByZXBseS4uLlxuICAgICAgZGVsZXRlIHJwY1Byb21pc2VzW25dXG4gICAgZWxzZSBpZiBlbWl0dGVyXG4gICAgICBpZiBtc2cgYW5kIHJlcGx5Lmxlbmd0aFxuICAgICAgICBlbWl0dGVyLmVtaXQgbXNnLCByZXBseVswXVxuICAgICAgZWxzZVxuICAgICAgICBkZWxldGUgcnBjUHJvbWlzZXNbbl0gIyBzdG9wIHN0cmVhbWluZ1xuICAgICAgICBlbWl0dGVyLmVtaXQgJ2Nsb3NlJywgcmVwbHlbMF1cbiAgICBlbHNlXG4gICAgICBjb25zb2xlLmVycm9yIFwic3B1cmlvdXMgcnBjIHJlcGx5XCIsIG4sIG1zZywgcmVwbHkuLi5cblxuICAjIFNldCB1cCBhIHdlYnNvY2tldCBjb25uZWN0aW9uIHRvIHRoZSBKZWVCdXMgc2VydmVyLlxuICAjIFRoZSBhcHBUYWcgaXMgdGhlIGRlZmF1bHQgdGFnIHRvIHVzZSB3aGVuIHNlbmRpbmcgcmVxdWVzdHMgdG8gaXQuXG4gIGNvbm5lY3QgPSAoYXBwVGFnKSAtPlxuXG4gICAgcmVjb25uZWN0ID0gKGZpcnN0Q2FsbCkgLT5cbiAgICAgICMgdGhlIHdlYnNvY2tldCBpcyBzZXJ2ZWQgZnJvbSB0aGUgc2FtZSBzaXRlIGFzIHRoZSB3ZWIgcGFnZVxuICAgICAgd3MgPSBuZXcgV2ViU29ja2V0IFwid3M6Ly8je2xvY2F0aW9uLmhvc3R9L3dzXCIsIFthcHBUYWddXG5cbiAgICAgIHdzLm9ub3BlbiA9IC0+XG4gICAgICAgICMgbG9jYXRpb24ucmVsb2FkKCkgIHVubGVzcyBmaXJzdENhbGxcbiAgICAgICAgY29uc29sZS5sb2cgJ1dTIE9wZW4nXG4gICAgICAgICRyb290U2NvcGUuJGFwcGx5IC0+XG4gICAgICAgICAgJHJvb3RTY29wZS4kYnJvYWRjYXN0ICd3cy1vcGVuJ1xuXG4gICAgICB3cy5vbm1lc3NhZ2UgPSAobSkgLT5cbiAgICAgICAgaWYgbS5kYXRhIGluc3RhbmNlb2YgQXJyYXlCdWZmZXJcbiAgICAgICAgICBjb25zb2xlLmxvZyAnYmluYXJ5IG1zZycsIG1cbiAgICAgICAgJHJvb3RTY29wZS4kYXBwbHkgLT5cbiAgICAgICAgICBkYXRhID0gSlNPTi5wYXJzZSBtLmRhdGFcbiAgICAgICAgICBzd2l0Y2ggdHlwZW9mIGRhdGFcbiAgICAgICAgICAgIHdoZW4gJ29iamVjdCdcbiAgICAgICAgICAgICAgaWYgQXJyYXkuaXNBcnJheSBkYXRhXG4gICAgICAgICAgICAgICAgcHJvY2Vzc1JwY1JlcGx5IGRhdGEuLi5cbiAgICAgICAgICAgICAgZWxzZVxuICAgICAgICAgICAgICAgIGNvbnNvbGUubG9nIFwic3B1cmlvdXMgb2JqZWN0IHJlY2VpdmVkXCI6IG1cbiAgICAgICAgICAgIHdoZW4gJ2Jvb2xlYW4nXG4gICAgICAgICAgICAgIGlmIGRhdGEgIyByZWxvYWQgYXBwXG4gICAgICAgICAgICAgICAgd2luZG93LmxvY2F0aW9uLnJlbG9hZCB0cnVlXG4gICAgICAgICAgICAgIGVsc2UgIyByZWZyZXNoIHN0eWxlc2hlZXRzXG4gICAgICAgICAgICAgICAgY29uc29sZS5sb2cgXCJDU1MgUmVsb2FkXCJcbiAgICAgICAgICAgICAgICBmb3IgZSBpbiBkb2N1bWVudC5nZXRFbGVtZW50c0J5VGFnTmFtZSAnbGluaydcbiAgICAgICAgICAgICAgICAgIGlmIGUuaHJlZiBhbmQgL3N0eWxlc2hlZXQvaS50ZXN0IGUucmVsXG4gICAgICAgICAgICAgICAgICAgIGUuaHJlZiA9IFwiI3tlLmhyZWYucmVwbGFjZSAvXFw/LiovLCAnJ30/I3tEYXRlLm5vdygpfVwiXG4gICAgICAgICAgICBlbHNlXG4gICAgICAgICAgICAgIGNvbnNvbGUubG9nICdTZXJ2ZXIgbXNnOicsIGRhdGFcblxuICAgICAgIyB3cy5vbmVycm9yID0gKGUpIC0+XG4gICAgICAjICAgY29uc29sZS5sb2cgJ0Vycm9yJywgZVxuXG4gICAgICB3cy5vbmNsb3NlID0gLT5cbiAgICAgICAgY29uc29sZS5sb2cgJ1dTIExvc3QnXG4gICAgICAgICRyb290U2NvcGUuJGFwcGx5IC0+XG4gICAgICAgICAgJHJvb3RTY29wZS4kYnJvYWRjYXN0ICd3cy1sb3N0J1xuICAgICAgICBzZXRUaW1lb3V0IHJlY29ubmVjdCwgMTAwMFxuXG4gICAgcmVjb25uZWN0IHRydWVcbiAgIFxuICAjIFNlbmQgYSBwYXlsb2FkIHRvIHRoZSBKZWVCdXMgc2VydmVyIG92ZXIgdGhlIHdlYnNvY2tldCBjb25uZWN0aW9uLlxuICAjIFRoZSBwYXlsb2FkIHNob3VsZCBiZSBhbiBvYmplY3QgKGFueXRoaW5nIGJ1dCBhcnJheSBpcyBzdXBwb3J0ZWQgZm9yIG5vdykuXG4gIHNlbmQgPSAocGF5bG9hZCkgLT5cbiAgICB3cy5zZW5kIGFuZ3VsYXIudG9Kc29uIHBheWxvYWRcbiAgICBAXG5cbiAgIyBGZXRjaCBhIGtleS92YWx1ZSBwYWlyIGZyb20gdGhlIHNlcnZlciBkYXRhYmFzZSwgdmFsdWUgcmV0dXJuZWQgYXMgcHJvbWlzZS5cbiAgZ2V0ID0gKGtleSkgLT5cbiAgICBycGMgJ2dldCcsIGtleVxuICAgICAgXG4gICMgU3RvcmUgYSBrZXkvdmFsdWUgcGFpciBpbiB0aGUgc2VydmVyIGRhdGFiYXNlLlxuICBwdXQgPSAoa2V5LCB2YWx1ZSkgLT5cbiAgICBzZW5kIFswLCAncHV0Jywga2V5LCB2YWx1ZV1cbiAgICBAXG4gICAgICBcbiAgIyBQZXJmb3JtIGFuIFJQQyBjYWxsLCBpLmUuIHJlZ2lzdGVyIHJlc3VsdCBjYWxsYmFjayBhbmQgcmV0dXJuIGEgcHJvbWlzZS5cbiAgcnBjID0gKGNtZCwgYXJncy4uLikgLT5cbiAgICBkID0gJHEuZGVmZXIoKVxuICAgIG4gPSArK3NlcU51bVxuICAgIHdzLnNlbmQgYW5ndWxhci50b0pzb24gW2NtZCwgbiwgYXJncy4uLl1cbiAgICB0ID0gc2V0VGltZW91dCAtPlxuICAgICAgY29uc29sZS5lcnJvciBcIlJQQyAje259OiBubyByZXBvbnNlXCIsIGFyZ3NcbiAgICAgIGRlbGV0ZSBycGNQcm9taXNlc1tuXVxuICAgICAgJHJvb3RTY29wZS4kYXBwbHkgLT5cbiAgICAgICAgZC5yZWplY3QoKVxuICAgICwgMTAwMDAgIyAxMCBzZWNvbmRzIHNob3VsZCBiZSBlbm91Z2ggdG8gY29tcGxldGUgYW55IHJlcXVlc3RcbiAgICBycGNQcm9taXNlc1tuXSA9IHRpbWVyOiB0LCBkZWZlcnJlZDogZFxuICAgIGQucHJvbWlzZVxuXG4gICMgTGF1bmNoIGEgZ2FkZ2V0IG9uIHRoZSBzZXJ2ZXIgYW5kIHJldHVybiBpdHMgcmVzdWx0cyB2aWEgZXZlbnRzLlxuICBnYWRnZXQgPSAoYXJncy4uLikgLT5cbiAgICBlID0gbmV3IEV2ZW50RW1pdHRlclxuICAgIHJwYyBhcmdzLi4uXG4gICAgICAudGhlbiAoZWVTZXR0ZXIpIC0+XG4gICAgICAgIGVlU2V0dGVyIGVcbiAgICBlXG4gIFxuICB3aW5kb3cuc2VuZCA9IHNlbmQgIyBjb25zb2xlIGFjY2VzcywgZm9yIGRlYnVnZ2luZ1xuICB7Y29ubmVjdCxzZW5kLGdldCxwdXQscnBjLGdhZGdldH1cbiJdfQ==
