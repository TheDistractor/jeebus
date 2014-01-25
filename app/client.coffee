ng = angular.module 'myApp', ['ui.router']

ng.constant 'jbName', 'blinker'

ng.run ($rootScope, jbName) ->
  ws = null

  # general utility, available anywhere, to send an object to the JeeBus server
  $rootScope.jbSend = (payload) ->
    ws.send JSON.stringify { T: "sv/#{ws.protocol}", P: payload }
  
  reconnect = (firstCall) ->
    # the websocket is served from the same site as the web page
    ws = new WebSocket "ws://#{location.host}/ws", [jbName]
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
    $scope.jbSend {button,value}
