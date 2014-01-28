ng = angular.module 'myApp', ['ui.router']

ng.run (jeebus) ->
  jeebus.connect 'blinker'

ng.controller 'MainCtrl', ($scope, jeebus) ->

  $scope.button = (button, value) ->
    jeebus.send {button,value}

  $scope.echoTest = ->
    jeebus.send "echoTest!" # send a test message to JB server's stdout
    jeebus.rpc('echo', 'Echo', 'me!').then (r) ->
      $scope.message = r
