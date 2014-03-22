ng = angular.module 'myApp'

console.log 'NG', angular.version.full

ng.config ($urlRouterProvider, $locationProvider) ->
  $urlRouterProvider.otherwise '/'
  $locationProvider.html5Mode true
  
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
  processRpcReply = (n, err, result) ->
    [t,d] = rpcPromises[n]
    if d
      clearTimeout t
      if err
        console.error err
        d.reject err
      else
        d.resolve result
    else
      console.error "spurious rpc reply", err, n, result

  # Set up a websocket connection to the JeeBus server.
  # The appTag is the default tag to use when sending requests to it.
  connect = (appTag) ->

    reconnect = (firstCall) ->
      # the websocket is served from the same site as the web page
      ws = new WebSocket "ws://#{location.host}/ws", [appTag]

      ws.onopen = ->
        # location.reload()  unless firstCall
        console.log 'WS Open'
        $rootScope.$apply ->
          $rootScope.serverStatus = 'connected'

      ws.onmessage = (m) ->
        if m.data instanceof ArrayBuffer
          console.log 'binary msg', m
        $rootScope.$apply ->
          data = JSON.parse m.data
          switch typeof data
            when 'object'
              if Array.isArray data
                processRpcReply data...
              else
                for k, v of data
                  processModelUpdate k, v
            when 'boolean'
              if data # reload app
                window.location.reload true
              else # refresh stylesheets
                console.log "CSS Reload"
                for e in document.getElementsByTagName 'link'
                  if e.href and /stylesheet/i.test e.rel
                    e.href = "#{e.href.replace /\?.*/, ''}?#{Date.now()}"
            else
              console.log 'Server msg:', data

      # ws.onerror = (e) ->
      #   console.log 'Error', e

      ws.onclose = ->
        console.log 'WS Closed'
        $rootScope.$apply ->
          $rootScope.serverStatus = 'disconnected'
        setTimeout reconnect, 1000

    reconnect true
   
  # Send a payload to the JeeBus server over the websocket connection.
  # The payload should be an object (anything but array is supported for now).
  send = (payload) ->
    ws.send angular.toJson payload
    @

  # Fetch a key/value pair from the server database, value returned as promise.
  get = (key) ->
    rpc 'get', key
      
  # Store a key/value pair in the server database.
  put = (key, value) ->
    send [0, 'put', key, value]
    @
      
  # Perform an RPC call, i.e. register result callback and return a promise.
  rpc = (cmd, args...) ->
    d = $q.defer()
    n = ++seqNum
    ws.send angular.toJson [cmd, n, args...]
    t = setTimeout ->
      console.error "RPC #{n}: no reponse", args
      delete rpcPromises[n]
      $rootScope.$apply ->
        d.reject()
    , 10000 # 10 seconds should be enough to complete any request
    rpcPromises[n] = [t, d]
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
    @

  window.send = send # console access, for debugging
  {connect,send,get,put,rpc,attach,detach}
