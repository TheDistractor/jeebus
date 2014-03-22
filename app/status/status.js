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
        var Tag, i, id, key, loc, ms, row, val, value, x, _i, _len, _ref, _results;
        _results = [];
        for (_i = 0, _len = items.length; _i < _len; _i++) {
          x = items[_i];
          Tag = x.Tag, (_ref = x.Msg, loc = _ref.loc, ms = _ref.ms, val = _ref.val);
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
                  id: id
                });
              }
              row = $scope.readings[i];
              row.value = value;
              _results1.push(row.date = $filter('date')(ms, "MM-dd HH:mm:ss"));
            }
            return _results1;
          })());
        }
        return _results;
      });
    });
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsic3RhdHVzLmNvZmZlZSJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUFBLE1BQUEsRUFBQTs7QUFBQSxFQUFBLEVBQUEsR0FBSyxPQUFPLENBQUMsTUFBUixDQUFlLE9BQWYsQ0FBTCxDQUFBOztBQUFBLEVBRUEsRUFBRSxDQUFDLE1BQUgsQ0FBVSxTQUFDLGNBQUQsRUFBaUIsY0FBakIsR0FBQTtBQUNSLElBQUEsY0FBYyxDQUFDLEtBQWYsQ0FBcUIsUUFBckIsRUFDRTtBQUFBLE1BQUEsR0FBQSxFQUFLLFNBQUw7QUFBQSxNQUNBLFdBQUEsRUFBYSxvQkFEYjtBQUFBLE1BRUEsVUFBQSxFQUFZLFlBRlo7S0FERixDQUFBLENBQUE7V0FJQSxjQUFjLENBQUMsR0FBZixDQUFtQixTQUFuQixFQUE4QixRQUE5QixFQUF3QyxFQUF4QyxFQUxRO0VBQUEsQ0FBVixDQUZBLENBQUE7O0FBQUEsRUFTQSxFQUFFLENBQUMsVUFBSCxDQUFjLFlBQWQsRUFBNEIsU0FBQyxNQUFELEVBQVMsT0FBVCxFQUFrQixNQUFsQixHQUFBO0FBQzFCLFFBQUEsV0FBQTtBQUFBLElBQUEsTUFBTSxDQUFDLFFBQVAsR0FBa0IsRUFBbEIsQ0FBQTtBQUFBLElBQ0EsV0FBQSxHQUFjLEVBRGQsQ0FBQTtXQU1BLE1BQU0sQ0FBQyxHQUFQLENBQVcsU0FBWCxFQUFzQixTQUFBLEdBQUE7YUFFcEIsTUFBTSxDQUFDLE1BQVAsQ0FBYyxTQUFkLEVBQXlCO0FBQUEsUUFBQSxLQUFBLEVBQU8sV0FBUDtBQUFBLFFBQW9CLElBQUEsRUFBTSxPQUExQjtPQUF6QixDQUNFLENBQUMsRUFESCxDQUNNLEtBRE4sRUFDYSxTQUFDLEtBQUQsR0FBQTtBQUNULFlBQUEsc0VBQUE7QUFBQTthQUFBLDRDQUFBO3dCQUFBO0FBQ0UsVUFBQyxRQUFBLEdBQUQsWUFBSyxLQUFLLFdBQUEsS0FBSSxVQUFBLElBQUcsV0FBQSxJQUFqQixDQUFBO0FBQUE7O0FBQ0E7aUJBQUEsVUFBQTsrQkFBQTtBQUNFLGNBQUEsRUFBQSxHQUFLLEVBQUEsR0FBRSxDQUFBLEdBQUcsQ0FBQyxLQUFKLENBQVUsQ0FBVixDQUFBLENBQUYsR0FBZ0IsS0FBaEIsR0FBb0IsR0FBekIsQ0FBQTtBQUFBLGNBQ0EsQ0FBQSxHQUFJLFdBQVksQ0FBQSxFQUFBLENBRGhCLENBQUE7QUFFQSxjQUFBLElBQU8sU0FBUDtBQUNFLGdCQUFBLENBQUEsR0FBSSxXQUFZLENBQUEsRUFBQSxDQUFaLEdBQWtCLE1BQU0sQ0FBQyxRQUFRLENBQUMsTUFBdEMsQ0FBQTtBQUFBLGdCQUNBLE1BQU0sQ0FBQyxRQUFRLENBQUMsSUFBaEIsQ0FDRTtBQUFBLGtCQUFBLEdBQUEsRUFBSyxHQUFMO0FBQUEsa0JBQVUsR0FBQSxFQUFLLEdBQWY7QUFBQSxrQkFBb0IsS0FBQSxFQUFPLEVBQTNCO0FBQUEsa0JBQStCLElBQUEsRUFBTSxFQUFyQztBQUFBLGtCQUF5QyxFQUFBLEVBQUksRUFBN0M7aUJBREYsQ0FEQSxDQURGO2VBRkE7QUFBQSxjQU1BLEdBQUEsR0FBTSxNQUFNLENBQUMsUUFBUyxDQUFBLENBQUEsQ0FOdEIsQ0FBQTtBQUFBLGNBT0EsR0FBRyxDQUFDLEtBQUosR0FBWSxLQVBaLENBQUE7QUFBQSw2QkFTQSxHQUFHLENBQUMsSUFBSixHQUFXLE9BQUEsQ0FBUSxNQUFSLENBQUEsQ0FBZ0IsRUFBaEIsRUFBb0IsZ0JBQXBCLEVBVFgsQ0FERjtBQUFBOztlQURBLENBREY7QUFBQTt3QkFEUztNQUFBLENBRGIsRUFGb0I7SUFBQSxDQUF0QixFQVAwQjtFQUFBLENBQTVCLENBVEEsQ0FBQTtBQUFBIiwic291cmNlc0NvbnRlbnQiOlsibmcgPSBhbmd1bGFyLm1vZHVsZSAnbXlBcHAnXG5cbm5nLmNvbmZpZyAoJHN0YXRlUHJvdmlkZXIsIG5hdmJhclByb3ZpZGVyKSAtPlxuICAkc3RhdGVQcm92aWRlci5zdGF0ZSAnc3RhdHVzJyxcbiAgICB1cmw6ICcvc3RhdHVzJ1xuICAgIHRlbXBsYXRlVXJsOiAnc3RhdHVzL3N0YXR1cy5odG1sJ1xuICAgIGNvbnRyb2xsZXI6ICdTdGF0dXNDdHJsJ1xuICBuYXZiYXJQcm92aWRlci5hZGQgJy9zdGF0dXMnLCAnU3RhdHVzJywgMzBcblxubmcuY29udHJvbGxlciAnU3RhdHVzQ3RybCcsICgkc2NvcGUsICRmaWx0ZXIsIGplZWJ1cykgLT5cbiAgJHNjb3BlLnJlYWRpbmdzID0gW11cbiAgcmVhZGluZ3NNYXAgPSB7fVxuXG4gICMgJHNjb3BlLmh3aWQgPSBqZWVidXMuYXR0YWNoICcvamVlYm9vdC9od2lkLydcbiAgIyAkc2NvcGUuJG9uICckZGVzdHJveScsIC0+IGplZWJ1cy5kZXRhY2ggJy9qZWVib290L2h3aWQvJ1xuXG4gICRzY29wZS4kb24gJ3dzLW9wZW4nLCAtPlxuICAgIFxuICAgIGplZWJ1cy5nYWRnZXQgJ01RVFRTdWInLCBUb3BpYzogJy9zZW5zb3IvIycsIFBvcnQ6ICc6MTg4MydcbiAgICAgIC5vbiAnT3V0JywgKGl0ZW1zKSAtPlxuICAgICAgICBmb3IgeCBpbiBpdGVtc1xuICAgICAgICAgIHtUYWcsTXNnOntsb2MsbXMsdmFsfX0gPSB4XG4gICAgICAgICAgZm9yIGtleSwgdmFsdWUgb2YgdmFsXG4gICAgICAgICAgICBpZCA9IFwiI3tUYWcuc2xpY2UoOCl9IC0gI3trZXl9XCJcbiAgICAgICAgICAgIGkgPSByZWFkaW5nc01hcFtpZF1cbiAgICAgICAgICAgIHVubGVzcyBpP1xuICAgICAgICAgICAgICBpID0gcmVhZGluZ3NNYXBbaWRdID0gJHNjb3BlLnJlYWRpbmdzLmxlbmd0aFxuICAgICAgICAgICAgICAkc2NvcGUucmVhZGluZ3MucHVzaFxuICAgICAgICAgICAgICAgIGxvYzogbG9jLCBrZXk6IGtleSwgdmFsdWU6IFwiXCIsIGRhdGU6IFwiXCIsIGlkOiBpZFxuICAgICAgICAgICAgcm93ID0gJHNjb3BlLnJlYWRpbmdzW2ldXG4gICAgICAgICAgICByb3cudmFsdWUgPSB2YWx1ZVxuICAgICAgICAgICAgIyBjb252ZXJ0aW5nIHRvIGEgc3RyaW5nIGhlcmUgYXBwZWFycyB0byBiZSBtb3JlIGVmZmljaWVudC4uLlxuICAgICAgICAgICAgcm93LmRhdGUgPSAkZmlsdGVyKCdkYXRlJykobXMsIFwiTU0tZGQgSEg6bW06c3NcIilcbiJdfQ==
