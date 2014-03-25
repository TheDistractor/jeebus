ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'data',
    url: '/data'
    templateUrl: 'data/data.html'
    controller: dataCtrl
  navbarProvider.add '/data', 'Data', 35

dataCtrl = ($scope, jeebus) ->
  $scope.info =
    driver: [
      { id: "id", name: "Parameter" }
      { id: "name", name: "Name" }
      { id: "unit", name: "Unit" }
      { id: "factor", name: "Factor" }
      { id: "scale", name: "Scale" }
    ]
  
  $scope.table = 'driver'  
  $scope.columns = $scope.info[$scope.table]
  $scope.allowDelete = false

  $scope.deleteRow = ->
    if $scope.allowDelete and $scope.cursor?
      $scope.allowDelete = false
      console.log 'DELETE', $scope.cursor

  $scope.editRow = (row) ->
    $scope.cursor = row
    
  setup = ->
    jeebus.attach 'table'
      .on 'sync', -> console.log @keys
    jeebus.attach 'driver'
      .on 'init', -> $scope.rows = @rows
      
  setup()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', setup
