ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'status',
    url: '/status'
    templateUrl: 'status/status.html'
    controller: 'StatusCtrl'
  navbarProvider.add '/status', 'Status', 30

ng.controller 'StatusCtrl', ($scope, jeebus) ->
  # TODO rewrite these example to use the "hm" service i.s.o. "jeebus"

  $scope.echoTest = ->
    jeebus.send "Testing!" # send a test message to JB server's stdout
    jeebus.rpc('echo', 'Echo', 'yes!').then (r) ->
      $scope.message = r

  $scope.dbGetTest = ->
    jeebus.rpc('db-get', '/jb/info').then (r) ->
      $scope.message = r

  $scope.dbKeysTest = ->
    jeebus.rpc('db-keys', '/jb/').then (r) ->
      $scope.message = r
