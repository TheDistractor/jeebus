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
    sensor: [
      { id: "id", name: "Id" }
      { id: "loc", name: "Location" }
      # { id: "val", name: "Values" }
      { id: "ms", name: "Timestamp" }
      { id: "typ", name: "Type" }
    ]
  
  $scope.tables = Object.keys $scope.info
  console.log 'ta', $scope.tables
  $scope.table = 'driver'
  # $scope.table = 'sensor'
  
  $scope.columns = $scope.info[$scope.table]

  $scope.editRow = (row) ->
    $scope.cursor = row ? {}
    
  jeebus.attach = (table, rowHandler) ->
    g = jeebus.gadget 'Attach', In: "/#{table}/"

    g.store = (key, row) ->
      row.id = key
      @keys[row.id] ?= @rows.length
      @rows[@keys[row.id]] = row

    g.on 'Out', (m) ->
      switch m.Tag
        when '<range>' then @emit 'init', table
        when '<sync>' then @emit 'sync', table
        else @emit 'data', m.Tag.slice(2 + table.length), m.Msg
    g.on 'data', rowHandler ? g.store

    g.rows = []
    g.keys = {}
    g
    
  attach = ->
    jeebus.attach 'table'
      .on 'sync', ->
        console.log @rows
      
    dataVec = $scope.rows = []
    dataMap = {}

    dataHandler = (tag, msg) ->
      msg.id = tag
      dataMap[msg.id] ?= dataVec.length
      dataVec[dataMap[msg.id]] = msg
  
    jeebus.gadget 'Attach', In: "/#{$scope.table}/"
      .on 'Out', (m) ->
        switch m.Tag
          when '<range>' then @emit 'init'
          when '<sync>' then @emit 'sync'
          else @emit 'data', m.Tag.slice(2 + $scope.table.length), m.Msg
      .on 'data', dataHandler

  attach()  if $scope.serverStatus is 'connected'
  $scope.$on 'ws-open', -> attach()
