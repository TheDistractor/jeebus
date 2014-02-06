ng = angular.module 'myApp', ['ui.router']

ng.run (jeebus) ->
  jeebus.connect 'blinker'

ng.controller 'MainCtrl', ($scope, $timeout, jeebus) ->
  # TODO this delay seems to be required to avoid an error with WS setup - why?
  $timeout ->
    $scope.admin = jeebus.attach '/admin/'
    $scope.$on '$destroy', -> jeebus.detach '/admin/'
    $scope.blinker = jeebus.attach '/blinker/'
    $scope.$on '$destroy', -> jeebus.detach '/blinker/'
  , 100

  $scope.button = (button, value) ->
    jeebus.send {button,value}

  $scope.echoTest = ->
    jeebus.send "echoTest!" # send a test message to JB server's stdout
    jeebus.rpc('echo', 'Echo', 'me!').then (r) ->
      $scope.message = r

  $scope.dbGetTest = ->
    jeebus.rpc('db-get', '/jeebus/started').then (r) ->
      $scope.message = r

  $scope.dbKeysTest = ->
    jeebus.rpc('db-keys', '/').then (r) ->
      $scope.message = r

  $scope.luaTest = ->
    jeebus.rpc('lua', 'demo', 'twice', 111).then (r) ->
      $scope.message = r
