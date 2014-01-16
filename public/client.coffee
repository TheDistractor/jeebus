ng = angular.module 'myApp', ['ui.router']

ws = null

ng.run ($rootScope) ->

	setCount = (data) ->
		$rootScope.$apply ->
			$rootScope.count = data

	reconnect = (firstCall) ->
		# the websocket is served from the same site as the web page
		ws = new WebSocket "ws://#{location.host}/ws"

		ws.onopen = ->
			location.reload()  unless firstCall
			console.log 'Open'

		ws.onmessage = (m) ->
			setCount JSON.parse(m.data)

		ws.onclose = ->
			console.log 'Closed'
			setCount null
			setTimeout reconnect, 1000
		
	reconnect true

ng.controller 'MainCtrl', ($scope) ->
	$scope.leds =
		redLed: true
		greenLed: false
	$scope.button = (b, v) ->
		console.log JSON.stringify {b, v}
		ws.send JSON.stringify [b, v]
