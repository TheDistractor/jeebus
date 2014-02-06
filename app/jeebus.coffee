ng = angular.module 'myApp'

# The "jeebus" service below is the same for all client-side applications.
# It lets angular connect to the JeeBus server and send/receive messages.
ng.factory 'jeebus', ($rootScope, $q) ->
  ws = null          # the websocket object, while open
  seqNum = 0         # unique sequence numbers for each RPC request
  rpcPromises = {}   # maps seqNum to a pending <timerId,promise> entry
  trackedModels = {} # keeps track of which paths have been attached

  # Update one or more of the tracked models with an incoming change.
  processModelUpdate = (key, value) ->
    for k, info of trackedModels
      if k is key.slice(0, k.length)
        suffix = key.slice(k.length)
        if value
          info.model[suffix] = value
        else
          delete info.model[suffix]
    console.error "spurious model update", key, value  unless suffix

  # Resolve or reject a pending rpc promise.
  processRpcReply = (n, result, err) ->
    if rpcPromises[n]
      [tid,d] = rpcPromises[n]
      clearTimeout tid
      if err
        console.error err
        d.reject err
      else
        d.resolve result
    else
      console.error "spurious rpc reply", n, result, err

  # Set up a websocket connection to the JeeBus server.
  # The appTag is the default tag to use when sending requests to it.
  connect = (appTag, port) ->
    port ?= location.port # the default port is the same as the HTTP server

    reconnect = (firstCall) ->
      # the websocket is served from the same site as the web page
      # ws = new WebSocket "ws://#{location.host}/ws"
      ws = new WebSocket "ws://#{location.hostname}:#{port}/ws", [appTag]

      ws.onopen = ->
        # location.reload()  unless firstCall
        console.log 'WS Open'

      ws.onmessage = (m) ->
        if m.data instanceof ArrayBuffer
          console.log 'binary msg', m
        $rootScope.$apply ->
          data = JSON.parse(m.data)
          if m.data[0] is '['
            processRpcReply data...
          else
            # TODO should not write into the root scope (or merge, perhaps?)
            for k, v of data
              # $rootScope[k] = v
              processModelUpdate k, v

      # ws.onerror = (e) ->
      #   console.log 'Error', e

      ws.onclose = ->
        console.log 'WS Closed'
        setTimeout reconnect, 1000

    reconnect true
   
  # Send a payload to the JeeBus server over the websocket connection.
  # The payload should be an object (anything but array is supported for now).
  # This becomes an MQTT message with topic "sv/<appTag>/ip-<addr:port>".
  send = (payload) ->
    msg = angular.toJson payload
    if msg[0] is '['
      console.error "payload can't be an array (#{payload.length} elements)"
    else
      ws.send msg
    @

  # Store a key/value pair in the JeeBus database (key must start with "/").
  store = (key, value) ->
    msg = angular.toJson [key, value]
    if msg.slice(0, 3) is '["/'
      ws.send msg
    else
      console.error 'key does not start with "/":', key
    @
      
  # Perform an RPC call, i.e. register result callback and return a promise.
  # This doesn't use MQTT to avoid additional round trips for frequent calls.
  rpc = (args...) ->
    d = $q.defer()
    n = ++seqNum
    ws.send angular.toJson [n, args...]
    tid = setTimeout ->
      console.error "RPC #{n}: no reponse", args
      delete rpcPromises[n]
      $rootScope.$apply ->
        d.reject()
    , 10000 # 10 seconds should be enough to complete any request
    rpcPromises[n] = [tid, d]
    d.promise

  # Attach, i.e. get corresponding data as a model which tracks all changes.
  attach = (path) ->
    info = trackedModels[path] ?= { model: {}, count: 0 }
    if info.count++ is 0
      rpc 'attach', path
        .then (r) ->
          for k, v of r
            processModelUpdate k, v
          console.log 'attach', path
    info.model

  # Undo the effects of attaching, i.e. stop following changes.
  detach = (path) ->
    if trackedModels[path] && --trackedModels[path].count <= 0
      delete trackedModels[path]
      rpc 'detach', path
        .then -> console.log 'detach', path

  {connect,send,store,rpc,attach,detach}
