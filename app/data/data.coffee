ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'data',
    url: '/data'
    templateUrl: 'data/data.html'
    controller: 'DataCtrl'
  navbarProvider.add '/data', 'Data', 35

ng.controller 'DataCtrl', ($scope, jeebus) ->
  $scope.table = 'driver'
  # $scope.table = 'sensor'
  
  $scope.info =
    driver: [
      { id: "id", name: "Parameter" }
      { id: "name", name: "Name" }
      { id: "unit", name: "Unit" }
      { id: "factor", name: "Factor" }
      { id: "scale", name: "Scale" }
    ]
    sensor: [
      { id: "id", name: "Id" }
      { id: "loc", name: "Location" }
      { id: "val", name: "Values" }
      { id: "ms", name: "Timestamp" }
      { id: "typ", name: "Type" }
    ]
  
  $scope.columns = $scope.info[$scope.table]

  dataHandler = (tag, msg) ->
    msg.id = tag
    dataMap[msg.id] ?= dataVec.length
    dataVec[dataMap[msg.id]] = msg
  
  $scope.editRow = (row) ->
    $scope.cursor = row ? {}
    
  dataVec = $scope.rows = []
  dataMap = {}

  attach = ->
    jeebus.gadget 'Attach', In: "/#{$scope.table}/"
      .on 'Out', (m) ->
        if m.Tag[0] isnt '<'
          dataHandler m.Tag.slice(2 + $scope.table.length), m.Msg

  attach()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', -> attach()
