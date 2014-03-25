ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'status',
    url: '/status'
    templateUrl: 'status/status.html'
    controller: 'StatusCtrl'
  navbarProvider.add '/status', 'Status', 30

ng.controller 'StatusCtrl', ($scope, jeebus) ->
  # jeebus.attach 'sensor', (key, row) -> ...
  
  readingHandler = (tag, msg) ->
    # loc: ... val: [c1:12,c2:34,...]
    {loc,ms,val,typ} = msg
    for key, raw of val
      id = "#{tag} - #{key}" # device id
      tid = "#{typ}/#{key}"  # type id
      readingMap[id] ?= readingVec.length
      readingVec[readingMap[id]] = adjust {loc,key,raw,ms,typ,id,tid}

  unitHandler = (tag, msg) ->
    msg.id = tag
    # name: unit: scale: ...
    unitMap[msg.id] ?= unitVec.length
    unitVec[unitMap[msg.id]] = msg
    # update existing readings
    adjust r  for r in readingVec
    
  adjust = (row) ->
    info = unitVec[unitMap[row.tid]]
    row.value = row.raw
    if info?
      row.key = info.name
      row.unit = info.unit
      # apply some scaling and formatting
      if info.factor
        row.value *= info.factor
      if info.scale < 0
        row.value *= Math.pow 10, -info.scale
      else if info.scale >= 0
        row.value /= Math.pow 10, info.scale
        row.value = row.value.toFixed info.scale
    row

  lookupMaps = {}
  readingVec = $scope.readings = []
  readingMap = {}
  unitVec = $scope.units = []
  unitMap = {}

  attach = ->
    jeebus.gadget 'Attach', In: '/sensor/'
      .on 'Out', (m) ->
        if m.Tag[0] isnt '<'
          readingHandler m.Tag.slice(8), m.Msg
    jeebus.gadget 'Attach', In: '/driver/'
      .on 'Out', (m) ->
        if m.Tag[0] isnt '<'
          unitHandler m.Tag.slice(8), m.Msg

  attach()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', -> attach()
