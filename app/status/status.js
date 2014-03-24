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
    var attach, map, vec;
    vec = $scope.readings = [];
    map = {};
    attach = function() {
      return jeebus.gadget('MQTTSub', {
        Topic: '/sensor/#'
      }).on('Out', function(msg) {
        var Tag, id, key, loc, ms, typ, val, value, _ref, _results;
        Tag = msg.Tag, (_ref = msg.Msg, loc = _ref.loc, ms = _ref.ms, val = _ref.val, typ = _ref.typ);
        _results = [];
        for (key in val) {
          value = val[key];
          id = "" + (Tag.slice(8)) + " - " + key;
          if (map[id] == null) {
            map[id] = vec.length;
          }
          _results.push(vec[map[id]] = {
            loc: loc,
            key: key,
            value: value,
            ms: ms,
            typ: typ,
            id: id
          });
        }
        return _results;
      });
    };
    if ($scope.serverStatus === 'connected') {
      attach();
    }
    return $scope.$on('ws-open', function() {
      return attach();
    });
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsic3RhdHVzLmNvZmZlZSJdLCJuYW1lcyI6W10sIm1hcHBpbmdzIjoiQUFBQTtBQUFBLE1BQUEsRUFBQTs7QUFBQSxFQUFBLEVBQUEsR0FBSyxPQUFPLENBQUMsTUFBUixDQUFlLE9BQWYsQ0FBTCxDQUFBOztBQUFBLEVBRUEsRUFBRSxDQUFDLE1BQUgsQ0FBVSxTQUFDLGNBQUQsRUFBaUIsY0FBakIsR0FBQTtBQUNSLElBQUEsY0FBYyxDQUFDLEtBQWYsQ0FBcUIsUUFBckIsRUFDRTtBQUFBLE1BQUEsR0FBQSxFQUFLLFNBQUw7QUFBQSxNQUNBLFdBQUEsRUFBYSxvQkFEYjtBQUFBLE1BRUEsVUFBQSxFQUFZLFlBRlo7S0FERixDQUFBLENBQUE7V0FJQSxjQUFjLENBQUMsR0FBZixDQUFtQixTQUFuQixFQUE4QixRQUE5QixFQUF3QyxFQUF4QyxFQUxRO0VBQUEsQ0FBVixDQUZBLENBQUE7O0FBQUEsRUFTQSxFQUFFLENBQUMsVUFBSCxDQUFjLFlBQWQsRUFBNEIsU0FBQyxNQUFELEVBQVMsTUFBVCxHQUFBO0FBQzFCLFFBQUEsZ0JBQUE7QUFBQSxJQUFBLEdBQUEsR0FBTSxNQUFNLENBQUMsUUFBUCxHQUFrQixFQUF4QixDQUFBO0FBQUEsSUFDQSxHQUFBLEdBQU0sRUFETixDQUFBO0FBQUEsSUFHQSxNQUFBLEdBQVMsU0FBQSxHQUFBO2FBQ1AsTUFBTSxDQUFDLE1BQVAsQ0FBYyxTQUFkLEVBQXlCO0FBQUEsUUFBQSxLQUFBLEVBQU8sV0FBUDtPQUF6QixDQUNFLENBQUMsRUFESCxDQUNNLEtBRE4sRUFDYSxTQUFDLEdBQUQsR0FBQTtBQUNULFlBQUEsc0RBQUE7QUFBQSxRQUFDLFVBQUEsR0FBRCxjQUFLLEtBQUssV0FBQSxLQUFJLFVBQUEsSUFBRyxXQUFBLEtBQUksV0FBQSxJQUFyQixDQUFBO0FBQ0E7YUFBQSxVQUFBOzJCQUFBO0FBQ0UsVUFBQSxFQUFBLEdBQUssRUFBQSxHQUFFLENBQUEsR0FBRyxDQUFDLEtBQUosQ0FBVSxDQUFWLENBQUEsQ0FBRixHQUFnQixLQUFoQixHQUFvQixHQUF6QixDQUFBOztZQUNBLEdBQUksQ0FBQSxFQUFBLElBQU8sR0FBRyxDQUFDO1dBRGY7QUFBQSx3QkFFQSxHQUFJLENBQUEsR0FBSSxDQUFBLEVBQUEsQ0FBSixDQUFKLEdBQWU7QUFBQSxZQUFDLEtBQUEsR0FBRDtBQUFBLFlBQUssS0FBQSxHQUFMO0FBQUEsWUFBUyxPQUFBLEtBQVQ7QUFBQSxZQUFlLElBQUEsRUFBZjtBQUFBLFlBQWtCLEtBQUEsR0FBbEI7QUFBQSxZQUFzQixJQUFBLEVBQXRCO1lBRmYsQ0FERjtBQUFBO3dCQUZTO01BQUEsQ0FEYixFQURPO0lBQUEsQ0FIVCxDQUFBO0FBWUEsSUFBQSxJQUFhLE1BQU0sQ0FBQyxZQUFQLEtBQXVCLFdBQXBDO0FBQUEsTUFBQSxNQUFBLENBQUEsQ0FBQSxDQUFBO0tBWkE7V0FhQSxNQUFNLENBQUMsR0FBUCxDQUFXLFNBQVgsRUFBc0IsU0FBQSxHQUFBO2FBQUcsTUFBQSxDQUFBLEVBQUg7SUFBQSxDQUF0QixFQWQwQjtFQUFBLENBQTVCLENBVEEsQ0FBQTtBQUFBIiwic291cmNlc0NvbnRlbnQiOlsibmcgPSBhbmd1bGFyLm1vZHVsZSAnbXlBcHAnXG5cbm5nLmNvbmZpZyAoJHN0YXRlUHJvdmlkZXIsIG5hdmJhclByb3ZpZGVyKSAtPlxuICAkc3RhdGVQcm92aWRlci5zdGF0ZSAnc3RhdHVzJyxcbiAgICB1cmw6ICcvc3RhdHVzJ1xuICAgIHRlbXBsYXRlVXJsOiAnc3RhdHVzL3N0YXR1cy5odG1sJ1xuICAgIGNvbnRyb2xsZXI6ICdTdGF0dXNDdHJsJ1xuICBuYXZiYXJQcm92aWRlci5hZGQgJy9zdGF0dXMnLCAnU3RhdHVzJywgMzBcblxubmcuY29udHJvbGxlciAnU3RhdHVzQ3RybCcsICgkc2NvcGUsIGplZWJ1cykgLT5cbiAgdmVjID0gJHNjb3BlLnJlYWRpbmdzID0gW11cbiAgbWFwID0ge31cblxuICBhdHRhY2ggPSAtPlxuICAgIGplZWJ1cy5nYWRnZXQgJ01RVFRTdWInLCBUb3BpYzogJy9zZW5zb3IvIydcbiAgICAgIC5vbiAnT3V0JywgKG1zZykgLT5cbiAgICAgICAge1RhZyxNc2c6e2xvYyxtcyx2YWwsdHlwfX0gPSBtc2dcbiAgICAgICAgZm9yIGtleSwgdmFsdWUgb2YgdmFsXG4gICAgICAgICAgaWQgPSBcIiN7VGFnLnNsaWNlKDgpfSAtICN7a2V5fVwiXG4gICAgICAgICAgbWFwW2lkXSA/PSB2ZWMubGVuZ3RoXG4gICAgICAgICAgdmVjW21hcFtpZF1dID0ge2xvYyxrZXksdmFsdWUsbXMsdHlwLGlkfVxuXG4gIGF0dGFjaCgpICBpZiAkc2NvcGUuc2VydmVyU3RhdHVzIGlzICdjb25uZWN0ZWQnXG4gICRzY29wZS4kb24gJ3dzLW9wZW4nLCAtPiBhdHRhY2goKVxuIl19
