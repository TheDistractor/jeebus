ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'demo',
    url: '/'
    templateUrl: 'demo/demo.html'
    controller: 'DemoCtrl'
  navbarProvider.add '/', 'Demo', 25

ng.controller 'DemoCtrl', ($scope, jeebus) ->

  $scope.echoTest = ->
    jeebus.send "echoTest!" # send a test message to JB server's stdout
    jeebus.rpc 'echo', 'Echo', 'me!'
      .then (r) ->
        $scope.message = r

  $scope.dbKeysTest = ->
    jeebus.rpc 'db-keys', '/sensor/'
      .then (r) ->
        $scope.message = r

  $scope.mqttTest = ->
    jeebus.gadget 'MQTTSub', Topic: '/sensor/#', Port: ':1883'
      .on 'Out', (r) ->
        $scope.message = r
