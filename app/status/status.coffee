ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'status',
    url: '/status'
    templateUrl: 'status/status.html'
    controller: 'StatusCtrl'
  navbarProvider.add '/status', 'Status', 30

ng.controller 'StatusCtrl', ($scope, jeebus) ->
  vec = $scope.readings = []
  map = {}

  attach = ->
    jeebus.gadget 'MQTTSub', Topic: '/sensor/#'
      .on 'Out', (msg) ->
        {Tag,Msg:{loc,ms,val,typ}} = msg
        for key, value of val
          id = "#{Tag.slice(8)} - #{key}"
          map[id] ?= vec.length
          vec[map[id]] = {loc,key,value,ms,typ,id}

  attach()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', -> attach()
