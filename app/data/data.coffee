ng = angular.module 'myApp'

ng.config ($stateProvider, navbarProvider) ->
  $stateProvider.state 'data',
    url: '/data'
    templateUrl: 'data/data.html'
    controller: dataCtrl
  navbarProvider.add '/data', 'Data', 35

dataCtrl = ($scope, jeebus) ->
  $scope.table = 'table'  
  
  $scope.changeTable = ->
    $scope.cursor = null
    setup()  if $scope.serverStatus is 'connected'

  $scope.editRow = (row) ->
    $scope.cursor = row
    
  $scope.deleteRow = ->
    if $scope.allowDelete and $scope.cursor?
      $scope.allowDelete = false
      console.log 'DELETE', $scope.cursor

  # FIXME: this gets called far too often, and there's no cleanup yet!
  setup = ->
    $scope.tables = jeebus.attach 'table'
      .on 'sync', ->
        $scope.colInfo = @get($scope.table).attr.split ' '
        $scope.columns = jeebus.attach "column/#{$scope.table}"
          .on 'sync', ->
            $scope.data = jeebus.attach($scope.table)
      
  setup()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', setup
