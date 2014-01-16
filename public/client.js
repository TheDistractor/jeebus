// Generated by CoffeeScript 1.6.3
(function() {
  var ng, ws;

  ng = angular.module('myApp', ['ui.router']);

  ws = null;

  ng.run(function($rootScope) {
    var reconnect, setCount;
    setCount = function(data) {
      return $rootScope.$apply(function() {
        return $rootScope.count = data;
      });
    };
    reconnect = function(firstCall) {
      ws = new WebSocket("ws://" + location.host + "/ws");
      ws.onopen = function() {
        if (!firstCall) {
          location.reload();
        }
        return console.log('Open');
      };
      ws.onmessage = function(m) {
        return setCount(JSON.parse(m.data));
      };
      return ws.onclose = function() {
        console.log('Closed');
        setCount(null);
        return setTimeout(reconnect, 1000);
      };
    };
    return reconnect(true);
  });

  ng.controller('MainCtrl', function($scope) {
    $scope.leds = {
      redLed: true,
      greenLed: false
    };
    return $scope.button = function(b, v) {
      console.log(JSON.stringify({
        b: b,
        v: v
      }));
      return ws.send(JSON.stringify([b, v]));
    };
  });

}).call(this);
