ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'demo',
    url: '/'
    templateUrl: 'demo/demo.html'
    controller: 'DemoCtrl'
  navbarProvider.add '/', 'Demo', 25

ng.controller 'DemoCtrl', ($scope, jeebus) ->
  # TODO rewrite these example to use the "hm" service i.s.o. "jeebus"

  $scope.echoTest = ->
    jeebus.send "echoTest!" # send a test message to JB server's stdout
    jeebus.rpc('echo', 'Echo', 'me!').then (r) ->
      $scope.message = r

  $scope.dbGetTest = ->
    jeebus.rpc('db-get', '/jb/info').then (r) ->
      $scope.message = r

  $scope.dbKeysTest = ->
    jeebus.rpc('db-keys', '/').then (r) ->
      $scope.message = r
