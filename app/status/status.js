(function() {
  var ng;

  ng = angular.module('myApp');

  ng.config(function($stateProvider, navbarProvider) {
    $stateProvider.state('status', {
      url: '/status',
      templateUrl: 'status/status.html',
      controller: 'StatusCtrl'
    });
    return navbarProvider.add('/status', 'Status', 30);
  });

  ng.controller('StatusCtrl', function($scope, jeebus) {
    var readings, readingsMap;
    readings = $scope.readings = [];
    readingsMap = {};
    return $scope.$on('ws-open', function() {
      return jeebus.gadget('MQTTSub', {
        Topic: '/sensor/#',
        Port: ':1883'
      }).on('Out', function(items) {
        var Tag, i, k, loc, ms, tag, v, val, x, _i, _len, _ref, _results;
        _results = [];
        for (_i = 0, _len = items.length; _i < _len; _i++) {
          x = items[_i];
          Tag = x.Tag, (_ref = x.Msg, loc = _ref.loc, ms = _ref.ms, val = _ref.val);
          tag = Tag.slice(8);
          _results.push((function() {
            var _results1;
            _results1 = [];
            for (k in val) {
              v = val[k];
              i = readingsMap[k];
              if (i == null) {
                i = readingsMap[k] = readings.length;
                readings.push({
                  loc: loc,
                  key: k,
                  val: "",
                  ms: ms,
                  tag: tag
                });
              }
              _results1.push(readings[i].val = v);
            }
            return _results1;
          })());
        }
        return _results;
      });
    });
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsic3RhdHVzLmNvZmZlZSJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUFBLE1BQUEsRUFBQTs7QUFBQSxFQUFBLEVBQUEsR0FBSyxPQUFPLENBQUMsTUFBUixDQUFlLE9BQWYsQ0FBTCxDQUFBOztBQUFBLEVBRUEsRUFBRSxDQUFDLE1BQUgsQ0FBVSxTQUFDLGNBQUQsRUFBaUIsY0FBakIsR0FBQTtBQUNSLElBQUEsY0FBYyxDQUFDLEtBQWYsQ0FBcUIsUUFBckIsRUFDRTtBQUFBLE1BQUEsR0FBQSxFQUFLLFNBQUw7QUFBQSxNQUNBLFdBQUEsRUFBYSxvQkFEYjtBQUFBLE1BRUEsVUFBQSxFQUFZLFlBRlo7S0FERixDQUFBLENBQUE7V0FJQSxjQUFjLENBQUMsR0FBZixDQUFtQixTQUFuQixFQUE4QixRQUE5QixFQUF3QyxFQUF4QyxFQUxRO0VBQUEsQ0FBVixDQUZBLENBQUE7O0FBQUEsRUFTQSxFQUFFLENBQUMsVUFBSCxDQUFjLFlBQWQsRUFBNEIsU0FBQyxNQUFELEVBQVMsTUFBVCxHQUFBO0FBQzFCLFFBQUEscUJBQUE7QUFBQSxJQUFBLFFBQUEsR0FBVyxNQUFNLENBQUMsUUFBUCxHQUFrQixFQUE3QixDQUFBO0FBQUEsSUFDQSxXQUFBLEdBQWMsRUFEZCxDQUFBO1dBTUEsTUFBTSxDQUFDLEdBQVAsQ0FBVyxTQUFYLEVBQXNCLFNBQUEsR0FBQTthQUVwQixNQUFNLENBQUMsTUFBUCxDQUFjLFNBQWQsRUFBeUI7QUFBQSxRQUFBLEtBQUEsRUFBTyxXQUFQO0FBQUEsUUFBb0IsSUFBQSxFQUFNLE9BQTFCO09BQXpCLENBQ0UsQ0FBQyxFQURILENBQ00sS0FETixFQUNhLFNBQUMsS0FBRCxHQUFBO0FBQ1QsWUFBQSw0REFBQTtBQUFBO2FBQUEsNENBQUE7d0JBQUE7QUFDRSxVQUFDLFFBQUEsR0FBRCxZQUFLLEtBQUssV0FBQSxLQUFJLFVBQUEsSUFBRyxXQUFBLElBQWpCLENBQUE7QUFBQSxVQUNBLEdBQUEsR0FBTSxHQUFHLENBQUMsS0FBSixDQUFVLENBQVYsQ0FETixDQUFBO0FBQUE7O0FBRUE7aUJBQUEsUUFBQTt5QkFBQTtBQUNFLGNBQUEsQ0FBQSxHQUFJLFdBQVksQ0FBQSxDQUFBLENBQWhCLENBQUE7QUFDQSxjQUFBLElBQU8sU0FBUDtBQUNFLGdCQUFBLENBQUEsR0FBSSxXQUFZLENBQUEsQ0FBQSxDQUFaLEdBQWlCLFFBQVEsQ0FBQyxNQUE5QixDQUFBO0FBQUEsZ0JBQ0EsUUFBUSxDQUFDLElBQVQsQ0FBYztBQUFBLGtCQUFBLEdBQUEsRUFBSyxHQUFMO0FBQUEsa0JBQVUsR0FBQSxFQUFLLENBQWY7QUFBQSxrQkFBa0IsR0FBQSxFQUFLLEVBQXZCO0FBQUEsa0JBQTJCLEVBQUEsRUFBSSxFQUEvQjtBQUFBLGtCQUFtQyxHQUFBLEVBQUssR0FBeEM7aUJBQWQsQ0FEQSxDQURGO2VBREE7QUFBQSw2QkFJQSxRQUFTLENBQUEsQ0FBQSxDQUFFLENBQUMsR0FBWixHQUFrQixFQUpsQixDQURGO0FBQUE7O2VBRkEsQ0FERjtBQUFBO3dCQURTO01BQUEsQ0FEYixFQUZvQjtJQUFBLENBQXRCLEVBUDBCO0VBQUEsQ0FBNUIsQ0FUQSxDQUFBO0FBQUEiLCJzb3VyY2VzQ29udGVudCI6WyJuZyA9IGFuZ3VsYXIubW9kdWxlICdteUFwcCdcblxubmcuY29uZmlnICgkc3RhdGVQcm92aWRlciwgbmF2YmFyUHJvdmlkZXIpIC0+XG4gICRzdGF0ZVByb3ZpZGVyLnN0YXRlICdzdGF0dXMnLFxuICAgIHVybDogJy9zdGF0dXMnXG4gICAgdGVtcGxhdGVVcmw6ICdzdGF0dXMvc3RhdHVzLmh0bWwnXG4gICAgY29udHJvbGxlcjogJ1N0YXR1c0N0cmwnXG4gIG5hdmJhclByb3ZpZGVyLmFkZCAnL3N0YXR1cycsICdTdGF0dXMnLCAzMFxuXG5uZy5jb250cm9sbGVyICdTdGF0dXNDdHJsJywgKCRzY29wZSwgamVlYnVzKSAtPlxuICByZWFkaW5ncyA9ICRzY29wZS5yZWFkaW5ncyA9IFtdXG4gIHJlYWRpbmdzTWFwID0ge31cblxuICAjICRzY29wZS5od2lkID0gamVlYnVzLmF0dGFjaCAnL2plZWJvb3QvaHdpZC8nXG4gICMgJHNjb3BlLiRvbiAnJGRlc3Ryb3knLCAtPiBqZWVidXMuZGV0YWNoICcvamVlYm9vdC9od2lkLydcblxuICAkc2NvcGUuJG9uICd3cy1vcGVuJywgLT5cbiAgICBcbiAgICBqZWVidXMuZ2FkZ2V0ICdNUVRUU3ViJywgVG9waWM6ICcvc2Vuc29yLyMnLCBQb3J0OiAnOjE4ODMnXG4gICAgICAub24gJ091dCcsIChpdGVtcykgLT5cbiAgICAgICAgZm9yIHggaW4gaXRlbXNcbiAgICAgICAgICB7VGFnLE1zZzp7bG9jLG1zLHZhbH19ID0geFxuICAgICAgICAgIHRhZyA9IFRhZy5zbGljZSg4KVxuICAgICAgICAgIGZvciBrLCB2IG9mIHZhbFxuICAgICAgICAgICAgaSA9IHJlYWRpbmdzTWFwW2tdXG4gICAgICAgICAgICB1bmxlc3MgaT9cbiAgICAgICAgICAgICAgaSA9IHJlYWRpbmdzTWFwW2tdID0gcmVhZGluZ3MubGVuZ3RoXG4gICAgICAgICAgICAgIHJlYWRpbmdzLnB1c2ggbG9jOiBsb2MsIGtleTogaywgdmFsOiBcIlwiLCBtczogbXMsIHRhZzogdGFnXG4gICAgICAgICAgICByZWFkaW5nc1tpXS52YWwgPSB2XG4iXX0=
