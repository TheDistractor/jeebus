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
  rpcPromises = {}   # maps seqNum to a pending {timer,deferred,emitter} entry

  # Resolve or reject a pending rpc promise. Also handle streamed results.
  processRpcReply = (n, msg, reply...) ->
    {timer,deferred,emitter} = rpcPromises[n]
    if deferred
      clearTimeout timer
      if msg is true # start streaming
        rpcPromises[n].deferred = null
        deferred.resolve (ee) ->
          rpcPromises[n].emitter = ee
        return
      if msg is "" and reply.length
        deferred.resolve reply[0]
      else if msg and reply.length == 0
        console.error msg
        deferred.reject msg
      else
        console.error "bad rpc reply", n, msg, reply...
      delete rpcPromises[n]
    else if emitter
      if msg and reply.length
        emitter.emit msg, reply[0]
      else
        delete rpcPromises[n] # stop streaming
        emitter.emit 'close', reply[0]
    else
      console.error "spurious rpc reply", n, msg, reply...

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
          $rootScope.$broadcast 'ws-open'

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
                console.log "spurious object received": m
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
        console.log 'WS Lost'
        $rootScope.$apply ->
          $rootScope.$broadcast 'ws-lost'
        setTimeout reconnect, 1000

    reconnect true
   
  # Send a payload to the JeeBus server over the websocket connection.
  # The payload should be an object (anything but array is supported for now).
  send = (payload) ->
    ws.send angular.toJson payload
    @

  # Return list of keys as promise, matching the given prefix in the database.
  keys = (prefix) -> rpc 'keys', prefix
      
  # Fetch a key/value pair from the server database, value returned as promise.
  get = (key) -> rpc 'get', key
      
  # Store a key/value pair in the server database.
  put = (key, value) -> send [0, 'put', key, value]
      
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
    rpcPromises[n] = timer: t, deferred: d
    d.promise

  # Launch a gadget on the server and return its results via events.
  gadget = (args...) ->
    e = new EventEmitter
    rpc args...
      .then (eeSetter) ->
        eeSetter e
    e
  
  attach = (table, rowHandler) ->
    g = gadget 'Attach', In: "/#{table}/"

    g.get = (key) ->
      @rows[@keys[key]]
      
    g.set = (key, row) -> # store changes locally
      row.id = key
      @keys[row.id] ?= @rows.length
      @rows[@keys[row.id]] = row
      
    g.put = (key, row) -> # store changes locally as well as on server
      @set key, row
      put "/#{table}/#{key}", row

    g.on 'Out', (m) ->
      switch m.Tag
        when '<range>' then @emit 'init', table
        when '<sync>' then @emit 'sync', table
        else @emit 'data', m.Tag.slice(2 + table.length), m.Msg
    g.on 'data', rowHandler ? g.set

    g.rows = []
    g.keys = {}
    g
    
  # window.send = send # console access, for debugging
  {connect,send,keys,get,put,rpc,gadget,attach}
