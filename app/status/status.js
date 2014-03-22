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

  ng.controller('StatusCtrl', function($scope, $filter, jeebus) {
    var readingsMap;
    $scope.readings = [];
    readingsMap = {};
    return $scope.$on('ws-open', function() {
      return jeebus.gadget('MQTTSub', {
        Topic: '/sensor/#',
        Port: ':1883'
      }).on('Out', function(items) {
        var Tag, i, id, key, loc, ms, row, typ, val, value, x, _i, _len, _ref, _results;
        _results = [];
        for (_i = 0, _len = items.length; _i < _len; _i++) {
          x = items[_i];
          Tag = x.Tag, (_ref = x.Msg, loc = _ref.loc, ms = _ref.ms, val = _ref.val, typ = _ref.typ);
          _results.push((function() {
            var _results1;
            _results1 = [];
            for (key in val) {
              value = val[key];
              id = "" + (Tag.slice(8)) + " - " + key;
              i = readingsMap[id];
              if (i == null) {
                i = readingsMap[id] = $scope.readings.length;
                $scope.readings.push({
                  loc: loc,
                  key: key,
                  value: "",
                  date: "",
                  typ: typ,
                  id: id
                });
              }
              row = $scope.readings[i];
              row.value = value;
              _results1.push(row.time = $filter('date')(ms, "MM-dd HH:mm:ss"));
            }
            return _results1;
          })());
        }
        return _results;
      });
    });
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsic3RhdHVzLmNvZmZlZSJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUFBLE1BQUEsRUFBQTs7QUFBQSxFQUFBLEVBQUEsR0FBSyxPQUFPLENBQUMsTUFBUixDQUFlLE9BQWYsQ0FBTCxDQUFBOztBQUFBLEVBRUEsRUFBRSxDQUFDLE1BQUgsQ0FBVSxTQUFDLGNBQUQsRUFBaUIsY0FBakIsR0FBQTtBQUNSLElBQUEsY0FBYyxDQUFDLEtBQWYsQ0FBcUIsUUFBckIsRUFDRTtBQUFBLE1BQUEsR0FBQSxFQUFLLFNBQUw7QUFBQSxNQUNBLFdBQUEsRUFBYSxvQkFEYjtBQUFBLE1BRUEsVUFBQSxFQUFZLFlBRlo7S0FERixDQUFBLENBQUE7V0FJQSxjQUFjLENBQUMsR0FBZixDQUFtQixTQUFuQixFQUE4QixRQUE5QixFQUF3QyxFQUF4QyxFQUxRO0VBQUEsQ0FBVixDQUZBLENBQUE7O0FBQUEsRUFTQSxFQUFFLENBQUMsVUFBSCxDQUFjLFlBQWQsRUFBNEIsU0FBQyxNQUFELEVBQVMsT0FBVCxFQUFrQixNQUFsQixHQUFBO0FBQzFCLFFBQUEsV0FBQTtBQUFBLElBQUEsTUFBTSxDQUFDLFFBQVAsR0FBa0IsRUFBbEIsQ0FBQTtBQUFBLElBQ0EsV0FBQSxHQUFjLEVBRGQsQ0FBQTtXQU1BLE1BQU0sQ0FBQyxHQUFQLENBQVcsU0FBWCxFQUFzQixTQUFBLEdBQUE7YUFFcEIsTUFBTSxDQUFDLE1BQVAsQ0FBYyxTQUFkLEVBQXlCO0FBQUEsUUFBQSxLQUFBLEVBQU8sV0FBUDtBQUFBLFFBQW9CLElBQUEsRUFBTSxPQUExQjtPQUF6QixDQUNFLENBQUMsRUFESCxDQUNNLEtBRE4sRUFDYSxTQUFDLEtBQUQsR0FBQTtBQUNULFlBQUEsMkVBQUE7QUFBQTthQUFBLDRDQUFBO3dCQUFBO0FBQ0UsVUFBQyxRQUFBLEdBQUQsWUFBSyxLQUFLLFdBQUEsS0FBSSxVQUFBLElBQUcsV0FBQSxLQUFJLFdBQUEsSUFBckIsQ0FBQTtBQUFBOztBQUNBO2lCQUFBLFVBQUE7K0JBQUE7QUFDRSxjQUFBLEVBQUEsR0FBSyxFQUFBLEdBQUUsQ0FBQSxHQUFHLENBQUMsS0FBSixDQUFVLENBQVYsQ0FBQSxDQUFGLEdBQWdCLEtBQWhCLEdBQW9CLEdBQXpCLENBQUE7QUFBQSxjQUNBLENBQUEsR0FBSSxXQUFZLENBQUEsRUFBQSxDQURoQixDQUFBO0FBRUEsY0FBQSxJQUFPLFNBQVA7QUFDRSxnQkFBQSxDQUFBLEdBQUksV0FBWSxDQUFBLEVBQUEsQ0FBWixHQUFrQixNQUFNLENBQUMsUUFBUSxDQUFDLE1BQXRDLENBQUE7QUFBQSxnQkFDQSxNQUFNLENBQUMsUUFBUSxDQUFDLElBQWhCLENBQ0U7QUFBQSxrQkFBQSxHQUFBLEVBQUssR0FBTDtBQUFBLGtCQUFVLEdBQUEsRUFBSyxHQUFmO0FBQUEsa0JBQW9CLEtBQUEsRUFBTyxFQUEzQjtBQUFBLGtCQUErQixJQUFBLEVBQU0sRUFBckM7QUFBQSxrQkFBeUMsR0FBQSxFQUFLLEdBQTlDO0FBQUEsa0JBQW1ELEVBQUEsRUFBSSxFQUF2RDtpQkFERixDQURBLENBREY7ZUFGQTtBQUFBLGNBTUEsR0FBQSxHQUFNLE1BQU0sQ0FBQyxRQUFTLENBQUEsQ0FBQSxDQU50QixDQUFBO0FBQUEsY0FPQSxHQUFHLENBQUMsS0FBSixHQUFZLEtBUFosQ0FBQTtBQUFBLDZCQVNBLEdBQUcsQ0FBQyxJQUFKLEdBQVcsT0FBQSxDQUFRLE1BQVIsQ0FBQSxDQUFnQixFQUFoQixFQUFvQixnQkFBcEIsRUFUWCxDQURGO0FBQUE7O2VBREEsQ0FERjtBQUFBO3dCQURTO01BQUEsQ0FEYixFQUZvQjtJQUFBLENBQXRCLEVBUDBCO0VBQUEsQ0FBNUIsQ0FUQSxDQUFBO0FBQUEiLCJzb3VyY2VzQ29udGVudCI6WyJuZyA9IGFuZ3VsYXIubW9kdWxlICdteUFwcCdcblxubmcuY29uZmlnICgkc3RhdGVQcm92aWRlciwgbmF2YmFyUHJvdmlkZXIpIC0+XG4gICRzdGF0ZVByb3ZpZGVyLnN0YXRlICdzdGF0dXMnLFxuICAgIHVybDogJy9zdGF0dXMnXG4gICAgdGVtcGxhdGVVcmw6ICdzdGF0dXMvc3RhdHVzLmh0bWwnXG4gICAgY29udHJvbGxlcjogJ1N0YXR1c0N0cmwnXG4gIG5hdmJhclByb3ZpZGVyLmFkZCAnL3N0YXR1cycsICdTdGF0dXMnLCAzMFxuXG5uZy5jb250cm9sbGVyICdTdGF0dXNDdHJsJywgKCRzY29wZSwgJGZpbHRlciwgamVlYnVzKSAtPlxuICAkc2NvcGUucmVhZGluZ3MgPSBbXVxuICByZWFkaW5nc01hcCA9IHt9XG5cbiAgIyAkc2NvcGUuaHdpZCA9IGplZWJ1cy5hdHRhY2ggJy9qZWVib290L2h3aWQvJ1xuICAjICRzY29wZS4kb24gJyRkZXN0cm95JywgLT4gamVlYnVzLmRldGFjaCAnL2plZWJvb3QvaHdpZC8nXG5cbiAgJHNjb3BlLiRvbiAnd3Mtb3BlbicsIC0+XG4gICAgXG4gICAgamVlYnVzLmdhZGdldCAnTVFUVFN1YicsIFRvcGljOiAnL3NlbnNvci8jJywgUG9ydDogJzoxODgzJ1xuICAgICAgLm9uICdPdXQnLCAoaXRlbXMpIC0+XG4gICAgICAgIGZvciB4IGluIGl0ZW1zXG4gICAgICAgICAge1RhZyxNc2c6e2xvYyxtcyx2YWwsdHlwfX0gPSB4XG4gICAgICAgICAgZm9yIGtleSwgdmFsdWUgb2YgdmFsXG4gICAgICAgICAgICBpZCA9IFwiI3tUYWcuc2xpY2UoOCl9IC0gI3trZXl9XCJcbiAgICAgICAgICAgIGkgPSByZWFkaW5nc01hcFtpZF1cbiAgICAgICAgICAgIHVubGVzcyBpP1xuICAgICAgICAgICAgICBpID0gcmVhZGluZ3NNYXBbaWRdID0gJHNjb3BlLnJlYWRpbmdzLmxlbmd0aFxuICAgICAgICAgICAgICAkc2NvcGUucmVhZGluZ3MucHVzaFxuICAgICAgICAgICAgICAgIGxvYzogbG9jLCBrZXk6IGtleSwgdmFsdWU6IFwiXCIsIGRhdGU6IFwiXCIsIHR5cDogdHlwLCBpZDogaWRcbiAgICAgICAgICAgIHJvdyA9ICRzY29wZS5yZWFkaW5nc1tpXVxuICAgICAgICAgICAgcm93LnZhbHVlID0gdmFsdWVcbiAgICAgICAgICAgICMgY29udmVydGluZyB0byBhIHN0cmluZyBoZXJlIGFwcGVhcnMgdG8gYmUgbW9yZSBlZmZpY2llbnQuLi5cbiAgICAgICAgICAgIHJvdy50aW1lID0gJGZpbHRlcignZGF0ZScpKG1zLCBcIk1NLWRkIEhIOm1tOnNzXCIpXG4iXX0=
