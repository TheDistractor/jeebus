ng = angular.module 'myApp', ['ui.router']

ws = null

ng.run ($rootScope) ->

  reconnect = (firstCall) ->
    # the websocket is served from the same site as the web page
    ws = new WebSocket "ws://#{location.host}/ws", ['blinker']
    ws.binaryType = 'arraybuffer'

    ws.onopen = ->
      location.reload()  unless firstCall
      console.log 'Open', ws

    ws.onmessage = (m) ->
      if m.data instanceof ArrayBuffer
        console.log 'binary msg', m
      $rootScope.$apply ->
        for k, v of JSON.parse(m.data)
          $rootScope[k] = v

    # ws.onerror = (e) ->
    #   console.log 'Error', e

    ws.onclose = ->
      console.log 'Closed'
      setTimeout reconnect, 1000
    
  reconnect true

ng.controller 'MainCtrl', ($scope) ->

  $scope.button = (button, value) ->
    # ws.send angular.toJson {dest: 'sv/blinker', button, value}
    ws.send JSON.stringify ['sv/blinker', {button,value}]
