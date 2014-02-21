// Generated by CoffeeScript 1.7.1
(function() {
  var ng;

  ng = angular.module('myApp');

  ng.provider('navbar', function() {
    var navs;
    navs = [];
    return {
      add: function(route, title, weight) {
        if (weight == null) {
          weight = 50;
        }
        return navs.push({
          route: route,
          title: title,
          weight: weight
        });
      },
      del: function(route) {
        return navs = navs.filter(function(x) {
          return x.route !== route;
        });
      },
      $get: function() {
        return navs.sort(function(a, b) {
          return a.weight - b.weight;
        });
      }
    };
  });

  ng.config(function($urlRouterProvider, $locationProvider) {
    $urlRouterProvider.otherwise('/');
    return $locationProvider.html5Mode(true);
  });

  ng.controller('NavCtrl', function($scope, navbar) {
    return $scope.navbar = navbar;
  });

}).call(this);

//# sourceMappingURL=nav.map