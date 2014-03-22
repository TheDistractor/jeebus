ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'status',
    url: '/status'
    templateUrl: 'status/status.html'
    controller: 'StatusCtrl'
  navbarProvider.add '/status', 'Status', 30

ng.controller 'StatusCtrl', ($scope, jeebus) ->
  $scope.readings = []
  readingsMap = {}

  # $scope.hwid = jeebus.attach '/jeeboot/hwid/'
  # $scope.$on '$destroy', -> jeebus.detach '/jeeboot/hwid/'

  $scope.$on 'ws-open', ->
    
    jeebus.gadget 'MQTTSub', Topic: '/sensor/#', Port: ':1883'
      .on 'Out', (items) ->
        for x in items
          {Tag,Msg:{loc,ms,val}} = x
          for k, v of val
            row = readingsMap[k]
            unless row?
              row = loc: loc, key: k, val: "", ms: "", tag: Tag.slice(8)
              readingsMap[k] = row
              $scope.readings.push row
            row.val = v
            row.ms = ms
