ng = angular.module 'myApp', ['ui.router']

ws = null

ng.run ($rootScope) ->

  reconnect = (firstCall) ->
    # the websocket is served from the same site as the web page
    ws = new WebSocket "ws://#{location.host}/ws", ['JeeBus']
    ws.binaryType = 'arraybuffer'

    ws.onopen = ->
      location.reload()  unless firstCall
      console.log 'Open', ws

    ws.onmessage = (m) ->
      if m.data instanceof ArrayBuffer
        console.log 'binary msg', m
      $rootScope.$apply ->
        msg = JSON.parse(m.data)
        $rootScope.$broadcast msg[0], msg.slice(1) | 0

    ws.onclose = ->
      console.log 'Closed'
      setTimeout reconnect, 1000
    
  reconnect true

ng.controller 'MainCtrl', ($scope) ->
  $scope.$on 'R', (e, v) -> $scope.redLed = v is 1
  $scope.$on 'G', (e, v) -> $scope.greenLed = v is 1
  $scope.$on 'C', (e, v) -> $scope.count = v

  $scope.button = (b, v) ->
    ws.send JSON.stringify ['>if/serial/blinker', "L#{b}#{v}"]
