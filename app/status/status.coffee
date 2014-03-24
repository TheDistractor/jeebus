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

  # $scope.readings = jeebus.attach '/sensor/'
  # $scope.$on '$destroy', -> jeebus.detach '/sensor/'

  attach = ->
    vec = $scope.readings
    map = readingsMap
    
    jeebus.gadget 'MQTTSub', Topic: '/sensor/#'
      .on 'Out', (msg) ->
        {Tag,Msg:{loc,ms,val,typ}} = msg
        for key, value of val
          id = "#{Tag.slice(8)} - #{key}"
          map[id] ?= vec.length
          vec[map[id]] = {loc,key,value,ms,typ,id}
          # unless map[id]?
          #   map[id] = vec.length
          #   vec.push {loc,key,value,ms,typ,id}
          # row = vec[map[id]]
          # row.value = value
          # row.ms = ms

  attach()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', -> attach()

# # TODO: quick test to see how it could work, this belongs in jeebus.coffee
# attach = (scope, name, prefix) ->
#   vec = scope[name] ?= []
#   map = scope[name+'Map'] ?= {}
# 
#   scope.$on 'ws-open', ->
#   
#     jeebus.gadget 'MQTTSub', Topic: prefix + '#'
#       .on 'Out', (msg) ->
#         {Tag,Msg:{loc,ms,val,typ}} = msg
#         for key, value of val
#           id = "#{Tag.slice(prefix.length)} - #{key}"
#           i = map[id]
#           unless i?
#             i = map[id] = vec.length
#             vec.push
#               loc: loc, key: key, value: "", date: "", typ: typ, id: id
#           row = vec[i]
#           row.value = value
#           row.time = ms
#   
# attach $scope, 'readings', '/sensor/'
