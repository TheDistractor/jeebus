(function() {
  var dataCtrl, ng;

  ng = angular.module('myApp');

  ng.config(function($stateProvider, navbarProvider) {
    $stateProvider.state('data', {
      url: '/data',
      templateUrl: 'data/data.html',
      controller: dataCtrl
    });
    return navbarProvider.add('/data', 'Data', 35);
  });

  dataCtrl = function($scope, jeebus) {
    var setup;
    $scope.table = 'table';
    $scope.changeTable = function() {
      $scope.cursor = null;
      if ($scope.serverStatus === 'connected') {
        return setup();
      }
    };
    $scope.editRow = function(row) {
      return $scope.cursor = row;
    };
    $scope.deleteRow = function() {
      if ($scope.allowDelete && ($scope.cursor != null)) {
        $scope.allowDelete = false;
        return console.log('DELETE', $scope.cursor);
      }
    };
    setup = function() {
      return $scope.tables = jeebus.attach('table').on('sync', function() {
        $scope.colInfo = this.get($scope.table).attr.split(' ');
        return $scope.columns = jeebus.attach("column/" + $scope.table).on('sync', function() {
          return $scope.data = jeebus.attach($scope.table);
        });
      });
    };
    if ($scope.serverStatus === 'connected') {
      setup();
    }
    return $scope.$on('ws-open', setup);
  };

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiZGF0YS5jb2ZmZWUiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBQUE7QUFBQSxNQUFBLFlBQUE7O0FBQUEsRUFBQSxFQUFBLEdBQUssT0FBTyxDQUFDLE1BQVIsQ0FBZSxPQUFmLENBQUwsQ0FBQTs7QUFBQSxFQUVBLEVBQUUsQ0FBQyxNQUFILENBQVUsU0FBQyxjQUFELEVBQWlCLGNBQWpCLEdBQUE7QUFDUixJQUFBLGNBQWMsQ0FBQyxLQUFmLENBQXFCLE1BQXJCLEVBQ0U7QUFBQSxNQUFBLEdBQUEsRUFBSyxPQUFMO0FBQUEsTUFDQSxXQUFBLEVBQWEsZ0JBRGI7QUFBQSxNQUVBLFVBQUEsRUFBWSxRQUZaO0tBREYsQ0FBQSxDQUFBO1dBSUEsY0FBYyxDQUFDLEdBQWYsQ0FBbUIsT0FBbkIsRUFBNEIsTUFBNUIsRUFBb0MsRUFBcEMsRUFMUTtFQUFBLENBQVYsQ0FGQSxDQUFBOztBQUFBLEVBU0EsUUFBQSxHQUFXLFNBQUMsTUFBRCxFQUFTLE1BQVQsR0FBQTtBQUNULFFBQUEsS0FBQTtBQUFBLElBQUEsTUFBTSxDQUFDLEtBQVAsR0FBZSxPQUFmLENBQUE7QUFBQSxJQUVBLE1BQU0sQ0FBQyxXQUFQLEdBQXFCLFNBQUEsR0FBQTtBQUNuQixNQUFBLE1BQU0sQ0FBQyxNQUFQLEdBQWdCLElBQWhCLENBQUE7QUFDQSxNQUFBLElBQVksTUFBTSxDQUFDLFlBQVAsS0FBdUIsV0FBbkM7ZUFBQSxLQUFBLENBQUEsRUFBQTtPQUZtQjtJQUFBLENBRnJCLENBQUE7QUFBQSxJQU1BLE1BQU0sQ0FBQyxPQUFQLEdBQWlCLFNBQUMsR0FBRCxHQUFBO2FBQ2YsTUFBTSxDQUFDLE1BQVAsR0FBZ0IsSUFERDtJQUFBLENBTmpCLENBQUE7QUFBQSxJQVNBLE1BQU0sQ0FBQyxTQUFQLEdBQW1CLFNBQUEsR0FBQTtBQUNqQixNQUFBLElBQUcsTUFBTSxDQUFDLFdBQVAsSUFBdUIsdUJBQTFCO0FBQ0UsUUFBQSxNQUFNLENBQUMsV0FBUCxHQUFxQixLQUFyQixDQUFBO2VBQ0EsT0FBTyxDQUFDLEdBQVIsQ0FBWSxRQUFaLEVBQXNCLE1BQU0sQ0FBQyxNQUE3QixFQUZGO09BRGlCO0lBQUEsQ0FUbkIsQ0FBQTtBQUFBLElBZUEsS0FBQSxHQUFRLFNBQUEsR0FBQTthQUNOLE1BQU0sQ0FBQyxNQUFQLEdBQWdCLE1BQU0sQ0FBQyxNQUFQLENBQWMsT0FBZCxDQUNkLENBQUMsRUFEYSxDQUNWLE1BRFUsRUFDRixTQUFBLEdBQUE7QUFDVixRQUFBLE1BQU0sQ0FBQyxPQUFQLEdBQWlCLElBQUMsQ0FBQSxHQUFELENBQUssTUFBTSxDQUFDLEtBQVosQ0FBa0IsQ0FBQyxJQUFJLENBQUMsS0FBeEIsQ0FBOEIsR0FBOUIsQ0FBakIsQ0FBQTtlQUNBLE1BQU0sQ0FBQyxPQUFQLEdBQWlCLE1BQU0sQ0FBQyxNQUFQLENBQWUsU0FBQSxHQUFRLE1BQU0sQ0FBQyxLQUE5QixDQUNmLENBQUMsRUFEYyxDQUNYLE1BRFcsRUFDSCxTQUFBLEdBQUE7aUJBQ1YsTUFBTSxDQUFDLElBQVAsR0FBYyxNQUFNLENBQUMsTUFBUCxDQUFjLE1BQU0sQ0FBQyxLQUFyQixFQURKO1FBQUEsQ0FERyxFQUZQO01BQUEsQ0FERSxFQURWO0lBQUEsQ0FmUixDQUFBO0FBdUJBLElBQUEsSUFBWSxNQUFNLENBQUMsWUFBUCxLQUF1QixXQUFuQztBQUFBLE1BQUEsS0FBQSxDQUFBLENBQUEsQ0FBQTtLQXZCQTtXQXdCQSxNQUFNLENBQUMsR0FBUCxDQUFXLFNBQVgsRUFBc0IsS0FBdEIsRUF6QlM7RUFBQSxDQVRYLENBQUE7QUFBQSIsInNvdXJjZXNDb250ZW50IjpbIm5nID0gYW5ndWxhci5tb2R1bGUgJ215QXBwJ1xuXG5uZy5jb25maWcgKCRzdGF0ZVByb3ZpZGVyLCBuYXZiYXJQcm92aWRlcikgLT5cbiAgJHN0YXRlUHJvdmlkZXIuc3RhdGUgJ2RhdGEnLFxuICAgIHVybDogJy9kYXRhJ1xuICAgIHRlbXBsYXRlVXJsOiAnZGF0YS9kYXRhLmh0bWwnXG4gICAgY29udHJvbGxlcjogZGF0YUN0cmxcbiAgbmF2YmFyUHJvdmlkZXIuYWRkICcvZGF0YScsICdEYXRhJywgMzVcblxuZGF0YUN0cmwgPSAoJHNjb3BlLCBqZWVidXMpIC0+XG4gICRzY29wZS50YWJsZSA9ICd0YWJsZScgIFxuICBcbiAgJHNjb3BlLmNoYW5nZVRhYmxlID0gLT5cbiAgICAkc2NvcGUuY3Vyc29yID0gbnVsbFxuICAgIHNldHVwKCkgIGlmICRzY29wZS5zZXJ2ZXJTdGF0dXMgaXMgJ2Nvbm5lY3RlZCdcblxuICAkc2NvcGUuZWRpdFJvdyA9IChyb3cpIC0+XG4gICAgJHNjb3BlLmN1cnNvciA9IHJvd1xuICAgIFxuICAkc2NvcGUuZGVsZXRlUm93ID0gLT5cbiAgICBpZiAkc2NvcGUuYWxsb3dEZWxldGUgYW5kICRzY29wZS5jdXJzb3I/XG4gICAgICAkc2NvcGUuYWxsb3dEZWxldGUgPSBmYWxzZVxuICAgICAgY29uc29sZS5sb2cgJ0RFTEVURScsICRzY29wZS5jdXJzb3JcblxuICAjIEZJWE1FOiB0aGlzIGdldHMgY2FsbGVkIGZhciB0b28gb2Z0ZW4sIGFuZCB0aGVyZSdzIG5vIGNsZWFudXAgeWV0IVxuICBzZXR1cCA9IC0+XG4gICAgJHNjb3BlLnRhYmxlcyA9IGplZWJ1cy5hdHRhY2ggJ3RhYmxlJ1xuICAgICAgLm9uICdzeW5jJywgLT5cbiAgICAgICAgJHNjb3BlLmNvbEluZm8gPSBAZ2V0KCRzY29wZS50YWJsZSkuYXR0ci5zcGxpdCAnICdcbiAgICAgICAgJHNjb3BlLmNvbHVtbnMgPSBqZWVidXMuYXR0YWNoIFwiY29sdW1uLyN7JHNjb3BlLnRhYmxlfVwiXG4gICAgICAgICAgLm9uICdzeW5jJywgLT5cbiAgICAgICAgICAgICRzY29wZS5kYXRhID0gamVlYnVzLmF0dGFjaCgkc2NvcGUudGFibGUpXG4gICAgICBcbiAgc2V0dXAoKSAgaWYgJHNjb3BlLnNlcnZlclN0YXR1cyBpcyAnY29ubmVjdGVkJ1xuICAkc2NvcGUuJG9uICd3cy1vcGVuJywgc2V0dXBcbiJdfQ==
