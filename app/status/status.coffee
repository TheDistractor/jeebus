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
      .on 'Out', (msg) ->
        # loc: ... val: [c1:12,c2:34,...]
        {Tag:dev,Msg:{loc,ms,val,typ}} = msg
        for key, raw of val
          did = "#{dev.slice(8)} - #{key}" # device id
          tid = "#{typ} - #{key}"          # type id
          readingMap[did] ?= readingVec.length
          readingVec[readingMap[did]] = update {loc,key,raw,ms,typ,did,tid}
          
    jeebus.gadget 'Attach', Topic: '/driver/'
      .on 'Out', (msg) ->
        # name: unit: scale: ...
        {Tag:tag,Msg:info} = msg
        tid = "#{Tag.slice(8)} - #{key}"
        unitMap[tid] ?= unitVec.length
        unitVec[unitMap[tid]] = info
        # update existing readings
        update row  for row in readingVec

  update = (row) ->
    info = unitVec[unitMap[row.tid]]
    if info?
      row.name = info.name
      row.unit = info.unit
    row.value = row.raw
    row

  attach()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', -> attach()
