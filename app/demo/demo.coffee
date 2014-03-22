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
    jeebus.rpc 'echo', 'Echo', 'me!'
      .then (r) ->
        $scope.message = r

  $scope.dbKeysTest = ->
    jeebus.rpc 'db-keys', '/jb/'
      .then (r) ->
        $scope.message = r

  $scope.clockTest = ->
    jeebus.gadget 'Clock', Rate: '5s'
      .on 'Out', (r) ->
        $scope.tick = r
