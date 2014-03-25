ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'status',
    url: '/status'
    templateUrl: 'status/status.html'
    controller: 'StatusCtrl'
  navbarProvider.add '/status', 'Status', 30

ng.controller 'StatusCtrl', ($scope, jeebus) ->
  readingVec = $scope.readings = []
  readingMap = {}
  unitVec = $scope.units = []
  unitMap = {}

  attach = ->
    jeebus.gadget 'Attach', In: '/sensor/'
      .on 'Out', (message) ->
        # loc: ... val: [c1:12,c2:34,...]
        {Tag:tag,Msg:msg} = message
        if tag[0] isnt '<'
          {loc,ms,val,typ} = msg
          for key, raw of val
            did = "#{tag.slice(8)} - #{key}" # device id
            tid = "#{typ}/#{key}"            # type id
            readingMap[did] ?= readingVec.length
            readingVec[readingMap[did]] = adjust {loc,key,raw,ms,typ,did,tid}
          
    jeebus.gadget 'Attach', In: '/driver/'
      .on 'Out', (message) ->
        # name: unit: scale: ...
        {Tag:tag,Msg:msg} = message
        if tag[0] isnt '<'
          tid = tag.slice(8)
          unitMap[tid] ?= unitVec.length
          unitVec[unitMap[tid]] = msg

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

  attach()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', -> attach()
