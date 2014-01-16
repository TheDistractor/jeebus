# the websocket is served from the same site as the web page

reconnect = ->
	ws = new WebSocket "ws://#{location.host}/ws"

	ws.onopen = ->
		console.log 'Open'

	ws.onmessage = (m) ->
		data = JSON.parse(m.data)
		console.log data

	ws.onclose = ->
		console.log 'Closed'
		setTimeout reconnect, 1000

reconnect()
