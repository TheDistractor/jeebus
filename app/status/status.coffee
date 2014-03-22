ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'status',
    url: '/status'
    templateUrl: 'status/status.html'
    controller: 'StatusCtrl'
  navbarProvider.add '/status', 'Status', 30

ng.controller 'StatusCtrl', ($scope, jeebus) ->
  readings = $scope.readings = []
  readingsMap = {}

  # $scope.hwid = jeebus.attach '/jeeboot/hwid/'
  # $scope.$on '$destroy', -> jeebus.detach '/jeeboot/hwid/'

  $scope.$on 'ws-open', ->
    
    jeebus.gadget 'MQTTSub', Topic: '/sensor/#', Port: ':1883'
      .on 'Out', (items) ->
        for x in items
          {Tag,Msg:{loc,ms,val}} = x
          tag = Tag.slice(8)
          for k, v of val
            i = readingsMap[k]
            unless i?
              i = readingsMap[k] = readings.length
              readings.push loc: loc, key: k, val: "", ms: ms, tag: tag
            readings[i].val = v
