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
    var readingsMap;
    $scope.readings = [];
    readingsMap = {};
    return $scope.$on('ws-open', function() {
      return jeebus.gadget('MQTTSub', {
        Topic: '/sensor/#',
        Port: ':1883'
      }).on('Out', function(items) {
        var Tag, k, loc, ms, row, v, val, x, _i, _len, _ref, _results;
        _results = [];
        for (_i = 0, _len = items.length; _i < _len; _i++) {
          x = items[_i];
          Tag = x.Tag, (_ref = x.Msg, loc = _ref.loc, ms = _ref.ms, val = _ref.val);
          _results.push((function() {
            var _results1;
            _results1 = [];
            for (k in val) {
              v = val[k];
              row = readingsMap[k];
              if (row == null) {
                row = {
                  loc: loc,
                  key: k,
                  val: "",
                  ms: "",
                  tag: Tag.slice(8)
                };
                readingsMap[k] = row;
                $scope.readings.push(row);
              }
              row.val = v;
              _results1.push(row.ms = ms);
            }
            return _results1;
          })());
        }
        return _results;
      });
    });
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsic3RhdHVzLmNvZmZlZSJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUFBLE1BQUEsRUFBQTs7QUFBQSxFQUFBLEVBQUEsR0FBSyxPQUFPLENBQUMsTUFBUixDQUFlLE9BQWYsQ0FBTCxDQUFBOztBQUFBLEVBRUEsRUFBRSxDQUFDLE1BQUgsQ0FBVSxTQUFDLGNBQUQsRUFBaUIsY0FBakIsR0FBQTtBQUNSLElBQUEsY0FBYyxDQUFDLEtBQWYsQ0FBcUIsUUFBckIsRUFDRTtBQUFBLE1BQUEsR0FBQSxFQUFLLFNBQUw7QUFBQSxNQUNBLFdBQUEsRUFBYSxvQkFEYjtBQUFBLE1BRUEsVUFBQSxFQUFZLFlBRlo7S0FERixDQUFBLENBQUE7V0FJQSxjQUFjLENBQUMsR0FBZixDQUFtQixTQUFuQixFQUE4QixRQUE5QixFQUF3QyxFQUF4QyxFQUxRO0VBQUEsQ0FBVixDQUZBLENBQUE7O0FBQUEsRUFTQSxFQUFFLENBQUMsVUFBSCxDQUFjLFlBQWQsRUFBNEIsU0FBQyxNQUFELEVBQVMsTUFBVCxHQUFBO0FBQzFCLFFBQUEsV0FBQTtBQUFBLElBQUEsTUFBTSxDQUFDLFFBQVAsR0FBa0IsRUFBbEIsQ0FBQTtBQUFBLElBQ0EsV0FBQSxHQUFjLEVBRGQsQ0FBQTtXQU1BLE1BQU0sQ0FBQyxHQUFQLENBQVcsU0FBWCxFQUFzQixTQUFBLEdBQUE7YUFFcEIsTUFBTSxDQUFDLE1BQVAsQ0FBYyxTQUFkLEVBQXlCO0FBQUEsUUFBQSxLQUFBLEVBQU8sV0FBUDtBQUFBLFFBQW9CLElBQUEsRUFBTSxPQUExQjtPQUF6QixDQUNFLENBQUMsRUFESCxDQUNNLEtBRE4sRUFDYSxTQUFDLEtBQUQsR0FBQTtBQUNULFlBQUEseURBQUE7QUFBQTthQUFBLDRDQUFBO3dCQUFBO0FBQ0UsVUFBQyxRQUFBLEdBQUQsWUFBSyxLQUFLLFdBQUEsS0FBSSxVQUFBLElBQUcsV0FBQSxJQUFqQixDQUFBO0FBQUE7O0FBQ0E7aUJBQUEsUUFBQTt5QkFBQTtBQUNFLGNBQUEsR0FBQSxHQUFNLFdBQVksQ0FBQSxDQUFBLENBQWxCLENBQUE7QUFDQSxjQUFBLElBQU8sV0FBUDtBQUNFLGdCQUFBLEdBQUEsR0FBTTtBQUFBLGtCQUFBLEdBQUEsRUFBSyxHQUFMO0FBQUEsa0JBQVUsR0FBQSxFQUFLLENBQWY7QUFBQSxrQkFBa0IsR0FBQSxFQUFLLEVBQXZCO0FBQUEsa0JBQTJCLEVBQUEsRUFBSSxFQUEvQjtBQUFBLGtCQUFtQyxHQUFBLEVBQUssR0FBRyxDQUFDLEtBQUosQ0FBVSxDQUFWLENBQXhDO2lCQUFOLENBQUE7QUFBQSxnQkFDQSxXQUFZLENBQUEsQ0FBQSxDQUFaLEdBQWlCLEdBRGpCLENBQUE7QUFBQSxnQkFFQSxNQUFNLENBQUMsUUFBUSxDQUFDLElBQWhCLENBQXFCLEdBQXJCLENBRkEsQ0FERjtlQURBO0FBQUEsY0FLQSxHQUFHLENBQUMsR0FBSixHQUFVLENBTFYsQ0FBQTtBQUFBLDZCQU1BLEdBQUcsQ0FBQyxFQUFKLEdBQVMsR0FOVCxDQURGO0FBQUE7O2VBREEsQ0FERjtBQUFBO3dCQURTO01BQUEsQ0FEYixFQUZvQjtJQUFBLENBQXRCLEVBUDBCO0VBQUEsQ0FBNUIsQ0FUQSxDQUFBO0FBQUEiLCJzb3VyY2VzQ29udGVudCI6WyJuZyA9IGFuZ3VsYXIubW9kdWxlICdteUFwcCdcblxubmcuY29uZmlnICgkc3RhdGVQcm92aWRlciwgbmF2YmFyUHJvdmlkZXIpIC0+XG4gICRzdGF0ZVByb3ZpZGVyLnN0YXRlICdzdGF0dXMnLFxuICAgIHVybDogJy9zdGF0dXMnXG4gICAgdGVtcGxhdGVVcmw6ICdzdGF0dXMvc3RhdHVzLmh0bWwnXG4gICAgY29udHJvbGxlcjogJ1N0YXR1c0N0cmwnXG4gIG5hdmJhclByb3ZpZGVyLmFkZCAnL3N0YXR1cycsICdTdGF0dXMnLCAzMFxuXG5uZy5jb250cm9sbGVyICdTdGF0dXNDdHJsJywgKCRzY29wZSwgamVlYnVzKSAtPlxuICAkc2NvcGUucmVhZGluZ3MgPSBbXVxuICByZWFkaW5nc01hcCA9IHt9XG5cbiAgIyAkc2NvcGUuaHdpZCA9IGplZWJ1cy5hdHRhY2ggJy9qZWVib290L2h3aWQvJ1xuICAjICRzY29wZS4kb24gJyRkZXN0cm95JywgLT4gamVlYnVzLmRldGFjaCAnL2plZWJvb3QvaHdpZC8nXG5cbiAgJHNjb3BlLiRvbiAnd3Mtb3BlbicsIC0+XG4gICAgXG4gICAgamVlYnVzLmdhZGdldCAnTVFUVFN1YicsIFRvcGljOiAnL3NlbnNvci8jJywgUG9ydDogJzoxODgzJ1xuICAgICAgLm9uICdPdXQnLCAoaXRlbXMpIC0+XG4gICAgICAgIGZvciB4IGluIGl0ZW1zXG4gICAgICAgICAge1RhZyxNc2c6e2xvYyxtcyx2YWx9fSA9IHhcbiAgICAgICAgICBmb3IgaywgdiBvZiB2YWxcbiAgICAgICAgICAgIHJvdyA9IHJlYWRpbmdzTWFwW2tdXG4gICAgICAgICAgICB1bmxlc3Mgcm93P1xuICAgICAgICAgICAgICByb3cgPSBsb2M6IGxvYywga2V5OiBrLCB2YWw6IFwiXCIsIG1zOiBcIlwiLCB0YWc6IFRhZy5zbGljZSg4KVxuICAgICAgICAgICAgICByZWFkaW5nc01hcFtrXSA9IHJvd1xuICAgICAgICAgICAgICAkc2NvcGUucmVhZGluZ3MucHVzaCByb3dcbiAgICAgICAgICAgIHJvdy52YWwgPSB2XG4gICAgICAgICAgICByb3cubXMgPSBtc1xuIl19
