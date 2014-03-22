ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'status',
    url: '/status'
    templateUrl: 'status/status.html'
    controller: 'StatusCtrl'
  navbarProvider.add '/status', 'Status', 30

ng.controller 'StatusCtrl', ($scope, $filter, jeebus) ->
  $scope.readings = []
  readingsMap = {}

  # $scope.hwid = jeebus.attach '/jeeboot/hwid/'
  # $scope.$on '$destroy', -> jeebus.detach '/jeeboot/hwid/'

  $scope.$on 'ws-open', ->
    
    jeebus.gadget 'MQTTSub', Topic: '/sensor/#', Port: ':1883'
      .on 'Out', (items) ->
        for x in items
          {Tag,Msg:{loc,ms,val}} = x
          for key, value of val
            id = "#{Tag.slice(8)} - #{key}"
            i = readingsMap[id]
            unless i?
              i = readingsMap[id] = $scope.readings.length
              $scope.readings.push
                loc: loc, key: key, value: "", date: "", id: id
            row = $scope.readings[i]
            row.value = value
            # converting to a string here appears to be more efficient...
            row.date = $filter('date')(ms, "MM-dd HH:mm:ss")
